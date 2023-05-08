package middleware

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

func UploadFile(next echo.HandlerFunc) echo.HandlerFunc {
   return func(c echo.Context) error {
      file, err := c.FormFile("image-project")

	  if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	  }

	  src, err := file.Open()

	  if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	  }

	  defer src.Close()

	  tempFile, err := ioutil.TempFile("uploads","image-*.png")

	  fmt.Println(tempFile.Name())

	  if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	  }

	  defer tempFile.Close()

	  if _, err := io.Copy(tempFile, src); err != nil {
          return c.JSON(http.StatusInternalServerError, err)
	  } 

	  data := tempFile.Name() 

	  fileName := data[8:] 

      c.Set("dataFile",fileName)

	  return next(c)
   }
}