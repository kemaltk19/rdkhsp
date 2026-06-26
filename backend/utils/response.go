package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type SuccessListResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

type SuccessSingleResponse struct {
	Data interface{} `json:"data"`
}

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// OK writes a 200 OK single resource JSON response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessSingleResponse{Data: data})
}

// Created writes a 201 Created single resource JSON response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessSingleResponse{Data: data})
}

// List writes a 200 OK list resource JSON response with meta pagination.
func List(c *gin.Context, data interface{}, page, limit int, total int64) {
	c.JSON(http.StatusOK, SuccessListResponse{
		Data: data,
		Meta: Meta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// Err writes an error JSON response with the appropriate HTTP status code.
func Err(c *gin.Context, status int, code, message string, fields map[string]string) {
	c.JSON(status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
	})
}
