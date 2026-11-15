package service

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
	"github.com/Reverseflash45/sipinjam-lab/app/repository"
	"github.com/Reverseflash45/sipinjam-lab/helper"
)

type RoomService struct {
	rooms repository.RoomRepository
}

func NewRoomService(rooms repository.RoomRepository) *RoomService {
	return &RoomService{rooms: rooms}
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
