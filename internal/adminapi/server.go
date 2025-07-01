package adminapi

import (
	"net/http"

	"bufio"
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
)

func RunServer(addr string) error {
	r := gin.Default()

	r.GET("/api/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Admin API is running",
		})
	})

	r.GET("/api/audit/logs", func(c *gin.Context) {
		const maxLines = 100
		file, err := os.Open("logs/audit.log")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open audit log"})
			return
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read audit log"})
			return
		}

		// Берём последние maxLines
		start := 0
		if len(lines) > maxLines {
			start = len(lines) - maxLines
		}
		lines = lines[start:]

		var events []map[string]interface{}
		for _, line := range lines {
			var evt map[string]interface{}
			if err := json.Unmarshal([]byte(line), &evt); err == nil {
				events = append(events, evt)
			}
		}

		c.JSON(http.StatusOK, gin.H{"events": events})
	})

	return r.Run(addr)
} 