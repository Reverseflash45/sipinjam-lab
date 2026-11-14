package service

import (
	"errors"

	"github.com/Reverseflash45/sipinjam-lab/app/repository"
	"github.com/Reverseflash45/sipinjam-lab/helper"
)

// translate mengubah error repository menjadi AppError yang dipahami client.
func translate(err error) error {
	var appErr *helper.AppError
	switch {
	case errors.As(err, &appErr):
		return appErr
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("DUPLICATE", "data dengan kode atau email tersebut sudah ada")
	case errors.Is(err, repository.ErrRoomConflict):
		return helper.Conflict("ROOM_CONFLICT", "ruang sudah dipinjam pada waktu tersebut")
	default:
		return helper.Internal(err)
	}
}
