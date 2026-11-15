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
	tx       repository.TxManager
}

func NewBookingService(
	bookings repository.BookingRepository,
	rooms repository.RoomRepository,
	items repository.ItemRepository,
	tx repository.TxManager,
) *BookingService {
	return &BookingService{
		bookings: bookings,
		rooms:    rooms,
		items:    items,
		tx:       tx,
	}
}

// POST /api/v1/bookings
func (s *BookingService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, err := helper.CurrentUser(c)
	if err != nil {
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
		return helper.Validation(map[string]string{
			"room_id": "pilih ruang atau minimal satu alat",
		})
	}

	if errs := ValidateWindow(req.StartAt, req.EndAt, time.Now()); errs != nil {
		return helper.Validation(errs)
	}

	limit := MaxActiveBookings(current.Role)
	if limit == 0 {
		return helper.Unauthorized("role tidak diizinkan mengajukan peminjaman")
	}

	active, err := s.bookings.CountActiveByUser(ctx, current.UserID)
	if err != nil {
		return translate(err)
	}

	if active >= limit {
		return helper.Conflict(
			"BOOKING_LIMIT",
			fmt.Sprintf("peminjaman aktif sudah mencapai batas (%d)", limit),
		)
	}

	if req.RoomID != nil {
		room, err := s.rooms.FindByID(ctx, *req.RoomID)
		if err != nil {
			return translate(err)
		}

		if !room.IsActive {
			return helper.Conflict("ROOM_INACTIVE", "ruang sedang tidak dapat dipinjam")
		}

		clash, err := s.bookings.FindApprovedByRoom(
			ctx, room.ID, req.StartAt, req.EndAt,
		)
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
			return helper.Conflict(
				"ITEM_UNAVAILABLE",
				fmt.Sprintf("%s hanya tersedia %d unit", item.Name, item.Stock),
			)
		}
	}

	booking := model.Booking{
		UserID:  current.UserID,
		RoomID:  req.RoomID,
		Purpose: req.Purpose,
		StartAt: req.StartAt,
		EndAt:   req.EndAt,
		Status:  model.StatusPending,
		Items:   req.Items,
	}

	var created *model.Booking

	err = s.tx.WithTx(ctx, func(txCtx context.Context, tx repository.DBTX) error {
		var createErr error
		created, createErr = s.bookings.Create(txCtx, tx, &booking)
		return createErr
	})

	if err != nil {
		return translate(err)
	}

	return helper.Created(
		c,
		"peminjaman diajukan, menunggu persetujuan laboran",
		created,
	)
}

// GET /api/v1/bookings/mine
func (s *BookingService) Mine(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, err := helper.CurrentUser(c)
	if err != nil {
		return helper.Unauthorized("belum terautentikasi")
	}

	list, err := s.bookings.ListByUser(ctx, current.UserID)
	if err != nil {
		return translate(err)
	}

	status := strings.TrimSpace(c.Query("status"))
	if status != "" {
		if !validBookingStatus(model.BookingStatus(status)) {
			return helper.BadRequest("status peminjaman tidak valid")
		}

		filtered := make([]model.Booking, 0)
		for _, booking := range list {
			if booking.Status == model.BookingStatus(status) {
				filtered = append(filtered, booking)
			}
		}
		list = filtered
	}

	return helper.Success(
		c, fiber.StatusOK,
		"riwayat peminjaman berhasil diambil",
		list,
	)
}

// GET /api/v1/bookings?status=pending
func (s *BookingService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var status *model.BookingStatus

	queryStatus := strings.TrimSpace(c.Query("status"))
	if queryStatus != "" {
		value := model.BookingStatus(queryStatus)
		if !validBookingStatus(value) {
			return helper.BadRequest("status peminjaman tidak valid")
		}
		status = &value
	}

	list, err := s.bookings.ListAll(ctx, status)
	if err != nil {
		return translate(err)
	}

	return helper.Success(
		c, fiber.StatusOK,
		"daftar peminjaman berhasil diambil",
		list,
	)
}

func validBookingStatus(status model.BookingStatus) bool {
	switch status {
	case model.StatusPending,
		model.StatusApproved,
		model.StatusRejected,
		model.StatusCancelled,
		model.StatusReturned:
		return true
	default:
		return false
	}
}

// GET /api/v1/bookings/:id
func (s *BookingService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, err := helper.CurrentUser(c)
	if err != nil {
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
		return helper.NotFound("data tidak ditemukan")
	}

	return helper.Success(
		c, fiber.StatusOK,
		"detail peminjaman berhasil diambil",
		b,
	)
}

func (s *BookingService) Approve(c *fiber.Ctx) error {
	return s.review(c, model.StatusApproved)
}

func (s *BookingService) Reject(c *fiber.Ctx) error {
	return s.review(c, model.StatusRejected)
}

func (s *BookingService) Cancel(c *fiber.Ctx) error {
	return s.review(c, model.StatusCancelled)
}

func (s *BookingService) Return(c *fiber.Ctx) error {
	return s.review(c, model.StatusReturned)
}

// review menjalankan perubahan status di dalam transaksi.
func (s *BookingService) review(c *fiber.Ctx, to model.BookingStatus) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, err := helper.CurrentUser(c)
	if err != nil {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req struct {
		Note string `json:"note"`
	}

	_ = c.BodyParser(&req)

	var note *string
	if n := strings.TrimSpace(req.Note); n != "" {
		note = &n
	}

	var result model.Booking

	err = s.tx.WithTx(ctx, func(txCtx context.Context, tx repository.DBTX) error {
		b, err := s.bookings.FindByIDForUpdate(txCtx, tx, id)
		if err != nil {
			return err
		}

		if to == model.StatusCancelled && b.UserID != current.UserID {
			return helper.NotFound("data tidak ditemukan")
		}

		if !CanTransition(b.Status, to) {
			return helper.Conflict(
				"INVALID_STATUS",
				fmt.Sprintf(
					"peminjaman berstatus %s tidak dapat diubah menjadi %s",
					b.Status, to,
				),
			)
		}

		if to == model.StatusApproved {
			if err := s.bookings.LockAndCheckItems(
				txCtx, tx, b.Items, b.StartAt, b.EndAt,
			); err != nil {
				return err
			}
		}

		var reviewer *int
		if to != model.StatusCancelled {
			reviewer = &current.UserID
		}

		var returnedAt *time.Time
		if to == model.StatusReturned {
			now := time.Now()
			returnedAt = &now
		}

		if err := s.bookings.UpdateStatus(
			txCtx, tx, id, to, reviewer, note, returnedAt,
		); err != nil {
			return err
		}

		b.Status = to
		b.ReviewedBy = reviewer
		b.ReviewNote = note
		b.ReturnedAt = returnedAt

		result = *b
		return nil
	})

	if err != nil {
		return translate(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"status peminjaman menjadi "+string(to),
		result,
	)
}
