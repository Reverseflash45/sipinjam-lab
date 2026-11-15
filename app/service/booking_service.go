package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
	"github.com/Reverseflash45/sipinjam-lab/app/repository"
	"github.com/Reverseflash45/sipinjam-lab/helper"
)

type BookingService struct {
	bookings repository.BookingRepository
	rooms    repository.RoomRepository
	items    repository.ItemRepository
	tx       *repository.TxManager
}

func NewBookingService(bookings repository.BookingRepository, rooms repository.RoomRepository,
	items repository.ItemRepository, tx *repository.TxManager) *BookingService {
	return &BookingService{bookings: bookings, rooms: rooms, items: items, tx: tx}
}

// POST /api/v1/bookings (mahasiswa, dosen)
func (s *BookingService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid (waktu dalam format RFC3339)")
	}
	req.Purpose = strings.TrimSpace(req.Purpose)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	if req.RoomID == nil && len(req.Items) == 0 {
		return helper.Validation(map[string]string{"room_id": "pilih ruang atau minimal satu alat"})
	}
	if errs := ValidateWindow(req.StartAt, req.EndAt, time.Now()); errs != nil {
		return helper.Validation(errs)
	}

	active, err := s.bookings.CountActiveByUser(ctx, current.UserID)
	if err != nil {
		return translate(err)
	}
	if limit := MaxActiveBookings(current.Role); active >= limit {
		return helper.Conflict("BOOKING_LIMIT",
			fmt.Sprintf("peminjaman aktif sudah mencapai batas (%d)", limit))
	}

	if req.RoomID != nil {
		room, err := s.rooms.FindByID(ctx, *req.RoomID)
		if err != nil {
			return translate(err)
		}
		if !room.IsActive {
			return helper.Conflict("ROOM_INACTIVE", "ruang sedang tidak dapat dipinjam")
		}
		// Pemeriksaan awal supaya peminjam langsung tahu. Penjaga yang
		// sebenarnya adalah EXCLUDE constraint saat peminjaman disetujui.
		clash, err := s.bookings.FindApprovedByRoom(ctx, room.ID, req.StartAt, req.EndAt)
		if err != nil {
			return translate(err)
		}
		if len(clash) > 0 {
			return translate(repository.ErrRoomConflict)
		}
	}
	for _, it := range req.Items {
		item, err := s.items.FindByID(ctx, it.ItemID)
		if err != nil {
			return translate(err)
		}
		if !item.IsActive || it.Quantity > item.Stock {
			return helper.Conflict("ITEM_UNAVAILABLE",
				fmt.Sprintf("%s hanya tersedia %d unit", item.Name, item.Stock))
		}
	}

	var created model.Booking
	err = s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		var err error
		created, err = s.bookings.WithTx(tx).Create(ctx, model.Booking{
			UserID: current.UserID, RoomID: req.RoomID, Purpose: req.Purpose,
			StartAt: req.StartAt, EndAt: req.EndAt, Items: req.Items,
		})
		return err
	})
	if err != nil {
		return translate(err)
	}
	return helper.Created(c, "peminjaman diajukan, menunggu persetujuan laboran", created)
}

// GET /api/v1/bookings/mine
func (s *BookingService) Mine(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	list, err := s.bookings.FindAll(ctx, model.BookingQuery{UserID: current.UserID, Status: c.Query("status")})
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "riwayat peminjaman berhasil diambil", list)
}

// GET /api/v1/bookings?status=pending (laboran, admin)
func (s *BookingService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.bookings.FindAll(ctx, model.BookingQuery{Status: c.Query("status")})
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar peminjaman berhasil diambil", list)
}

// GET /api/v1/bookings/:id
func (s *BookingService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	b, err := s.bookings.FindByID(ctx, id)
	if err != nil {
		return translate(err)
	}
	if !CanViewBooking(current, b.UserID) {
		// 404, bukan 403: keberadaan peminjaman orang lain tidak dibocorkan.
		return helper.NotFound("data tidak ditemukan")
	}
	return helper.Success(c, fiber.StatusOK, "detail peminjaman berhasil diambil", b)
}

// PATCH /api/v1/bookings/:id/approve (laboran, admin)
func (s *BookingService) Approve(c *fiber.Ctx) error { return s.review(c, model.StatusApproved) }

// PATCH /api/v1/bookings/:id/reject (laboran, admin)
func (s *BookingService) Reject(c *fiber.Ctx) error { return s.review(c, model.StatusRejected) }

// PATCH /api/v1/bookings/:id/cancel (pemilik)
func (s *BookingService) Cancel(c *fiber.Ctx) error { return s.review(c, model.StatusCancelled) }

// PATCH /api/v1/bookings/:id/return (laboran, admin)
func (s *BookingService) Return(c *fiber.Ctx) error { return s.review(c, model.StatusReturned) }

// review menjalankan satu perpindahan status di dalam transaction.
func (s *BookingService) review(c *fiber.Ctx, to string) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.ReviewRequest
	_ = c.BodyParser(&req) // catatan bersifat opsional
	var note *string
	if n := strings.TrimSpace(req.Note); n != "" {
		note = &n
	}

	var result model.Booking
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		bookings := s.bookings.WithTx(tx)

		b, err := bookings.FindByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if to == model.StatusCancelled && b.UserID != current.UserID {
			return helper.NotFound("data tidak ditemukan")
		}
		if !CanTransition(b.Status, to) {
			return helper.Conflict("INVALID_STATUS",
				fmt.Sprintf("peminjaman berstatus %s tidak dapat diubah menjadi %s", b.Status, to))
		}
		if to == model.StatusApproved {
			if err := s.ensureStock(ctx, tx, b); err != nil {
				return err
			}
		}

		if to == model.StatusReturned {
			result, err = bookings.MarkReturned(ctx, id)
		} else {
			var reviewer *int
			if to != model.StatusCancelled {
				reviewer = &current.UserID
			}
			result, err = bookings.UpdateStatus(ctx, id, to, reviewer, note)
		}
		result.Items = b.Items
		return err
	})
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "status peminjaman menjadi "+to, result)
}

// ensureStock memastikan setiap alat masih cukup pada rentang waktu
// peminjaman. Baris alat dikunci dulu supaya dua persetujuan bersamaan
// tidak sama-sama lolos.
func (s *BookingService) ensureStock(ctx context.Context, tx repository.DBTX, b model.Booking) error {
	if len(b.Items) == 0 {
		return nil
	}
	items := s.items.WithTx(tx)
	ids := make([]int, 0, len(b.Items))
	for _, it := range b.Items {
		ids = append(ids, it.ItemID)
	}
	if err := items.LockForUpdate(ctx, ids); err != nil {
		return err
	}
	for _, it := range b.Items {
		item, err := items.FindByID(ctx, it.ItemID)
		if err != nil {
			return err
		}
		borrowed, err := items.BorrowedQuantity(ctx, it.ItemID, b.StartAt, b.EndAt)
		if err != nil {
			return err
		}
		if borrowed+it.Quantity > item.Stock {
			return helper.Conflict("ITEM_UNAVAILABLE",
				fmt.Sprintf("%s tersisa %d unit pada waktu tersebut", item.Name, item.Stock-borrowed))
		}
	}
	return nil
}
