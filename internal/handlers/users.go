package handlers

import (
	"net/http"
)

type Users struct{}

func NewUsers() *Users { return &Users{} }

func (u *Users) Me(w http.ResponseWriter, r *http.Request) {
	WriteError(w, NewError("UNAUTHORIZED", "login required", http.StatusUnauthorized))
}
