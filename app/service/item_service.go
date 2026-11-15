package service

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
	"github.com/Reverseflash45/sipinjam-lab/app/repository"
	"github.com/Reverseflash45/sipinjam-lab/helper"
)

type ItemService struct {
	items repository.ItemRepository
}

func NewItemService(items repository.ItemRepository) *ItemService {
	return &ItemService{items: items}
}

// GET /api/v1/items
func (s *ItemService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	items, err := s.items.FindAll(ctx)
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar alat berhasil diambil", items)
}

// GET /api/v1/items/:id
func (s *ItemService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id harus berupa angka positif")
	}
	item, err := s.items.FindByID(ctx, id)
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "alat ditemukan", item)
}

// POST /api/v1/items (admin, laboran)
func (s *ItemService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.ItemRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	item, err := s.items.Create(ctx, req)
	if err != nil {
		return translate(err)
	}
	return helper.Created(c, "alat berhasil ditambahkan", item)
}

// PUT /api/v1/items/:id (admin, laboran)
func (s *ItemService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.ItemRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	item, err := s.items.Update(ctx, id, req)
	if err != nil {
		return translate(err)
	}
	return helper.Success(c, fiber.StatusOK, "alat berhasil diubah", item)
}
