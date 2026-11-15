package service

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
	"github.com/Reverseflash45/sipinjam-lab/app/repository"
	"github.com/Reverseflash45/sipinjam-lab/helper"
)

type RoomService struct {
	rooms    repository.RoomRepository
	bookings repository.BookingRepository
}

func NewRoomService(rooms repository.RoomRepository, bookings repository.BookingRepository) *RoomService {
	return &RoomService{rooms: rooms, bookings: bookings}
}

// GET /api/v1/rooms
func (s *RoomService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	rooms, err := s.rooms.FindAll(ctx)
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar ruang berhasil diambil", rooms)
}

// GET /api/v1/rooms/:id
func (s *RoomService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id harus berupa angka positif")
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "ruang ditemukan", room)
}

// GET /api/v1/rooms/:id/schedule?date=2026-11-20
// Jadwal yang sudah disetujui pada satu hari, untuk membantu peminjam
// memilih jam yang masih kosong.
func (s *RoomService) Schedule(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id harus berupa angka positif")
	}
	day, err := time.ParseInLocation("2006-01-02", c.Query("date", time.Now().Format("2006-01-02")), time.Local)
	if err != nil {
		return helper.BadRequest("format date harus YYYY-MM-DD")
	}
	if _, err := s.rooms.FindByID(ctx, id); err != nil {
		return translate(err)
	}
	list, err := s.bookings.FindApprovedByRoom(ctx, id, day, day.AddDate(0, 0, 1))
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "jadwal ruang berhasil diambil", list)
}

// POST /api/v1/rooms (admin)
func (s *RoomService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	room, err := s.rooms.Create(ctx, req)
	if err != nil {
		return translate(err)
	}
	return helper.Created(c, "ruang berhasil ditambahkan", room)
}

// PUT /api/v1/rooms/:id (admin)
func (s *RoomService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.RoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	room, err := s.rooms.Update(ctx, id, req)
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "ruang berhasil diubah", room)
}
