package handlers

import (
	"errors"
	"io" // Importante: necesario para leer los bytes del archivo
	"net/http"
	"strconv"

	"go-hexagonal-api/internal/core/domain"
	"go-hexagonal-api/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type HTTPUserHandler struct {
	service ports.UserService
}

func NewHTTPUserHandler(service ports.UserService) *HTTPUserHandler {
	return &HTTPUserHandler{service: service}
}

func (h *HTTPUserHandler) Register(c *gin.Context) {
	var input domain.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Register(input)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *HTTPUserHandler) Login(c *gin.Context) {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *HTTPUserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	user, err := h.service.GetUserByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *HTTPUserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var input domain.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(id, input)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *HTTPUserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	err = h.service.DeleteUser(id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado exitosamente"})
}

// Nuevo endpoint para la subida de archivos
func (h *HTTPUserHandler) UploadFile(c *gin.Context) {
	// 1. Obtener el archivo del form-data usando Gin (la key en Postman/Android debe ser "file")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se encontró ningún archivo en la petición o la llave no es 'file'"})
		return
	}

	// 2. Abrir el archivo subido
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo abrir el archivo procesado"})
		return
	}
	defer f.Close() // Es vital cerrar el archivo al terminar

	// 3. Leer los bytes del archivo usando el paquete 'io'
	fileBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer los bytes del archivo"})
		return
	}

	// 4. Pasar los bytes y el nombre original a la capa de servicios (nuestro nuevo método)
	fileURL, err := h.service.UploadFile(fileBytes, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5. Retornar la respuesta exitosa con la URL
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": file.Filename,
		"message":  "Archivo subido correctamente",
	})
}