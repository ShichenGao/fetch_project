package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PurchaseItem struct {
	ShortDesc string `json:"shortDescription" binding:"required"`
	Price     string `json:"price" binding:"required"`
}

type Receipt struct {
	Retailer     string         `json:"retailer" binding:"required"`
	PurchaseDate string         `json:"purchaseDate" binding:"required"`
	PurchaseTime string         `json:"purchaseTime" binding:"required"`
	Items        []PurchaseItem `json:"items"`
	Total        string         `json:"total" binding:"required"`
}

func calLenOfChar(str string) int {
	letterCount := 0
	for _, char := range str {
		if unicode.IsLetter(char) {
			letterCount++
		}
	}
	return letterCount
}

func isInt(num float64) bool {
	return num-math.Floor(num) == 0
}

func extractDayFromDate(str string) int {
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return 0
	}
	day := t.Day()
	return day
}

func extractHourFromTime(str string) int {
	parts := strings.Split(str, ":")
	if len(parts) > 0 {
		num, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			fmt.Println("err:", err)
			return 0
		}
		return int(num)
	}
	return 0
}

func calPoint(receipt *Receipt) int {
	retailerNamePoint := calLenOfChar(receipt.Retailer)
	itemPoint := len(receipt.Items) / 2 * 5
	totalPoint := 0
	total, err := strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		fmt.Println("err:", err)
		return 0
	}
	if isInt(total) {
		totalPoint += 50
	}
	if isInt(total / 0.25) {
		totalPoint += 25
	}
	itemDescPoint := 0.0
	for _, item := range receipt.Items {
		desc := item.ShortDesc
		trimDesc := strings.Trim(desc, " ")
		if (strings.Count(trimDesc, "")-1)%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				fmt.Println("err:", err)
				return 0
			}
			itemDescPoint += math.Ceil(price * 0.2)
		}
	}
	datePoint := 0
	day := extractDayFromDate(receipt.PurchaseDate)
	if err != nil {
		fmt.Println("err:", err)
		return 0
	}
	if day%2 != 0 {
		datePoint = 6
	}
	hourPoint := 0
	hour := extractHourFromTime(receipt.PurchaseTime)
	if hour >= 14 && hour < 16 {
		hourPoint = 10
	}
	return hourPoint + datePoint + itemPoint + totalPoint + retailerNamePoint + int(itemDescPoint)

}

func main() {
	r := gin.Default()
	idPointMap := make(map[string]int)
	r.POST("/receipts/process", func(c *gin.Context) {
		p := &Receipt{}
		if err := c.ShouldBind(p); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
			return
		}
		uuid := uuid.New()
		point := calPoint(p)
		idPointMap[uuid.String()] = point
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"id": uuid,
		})
	})
	r.GET("/receipts/:id/points", func(c *gin.Context) {
		id := c.Param("id")
		point := idPointMap[id]
		c.JSON(http.StatusOK, gin.H{
			"points": point,
		})
	})

	r.Run(":8080")

}
