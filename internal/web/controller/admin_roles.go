package controller
import (
    "github.com/gin-gonic/gin"
)
type AdminRoleController struct {}
func NewAdminRoleController(g *gin.RouterGroup) *AdminRoleController {
    a := &AdminRoleController{}
    g.GET("/list", a.list)
    return a
}
func (a *AdminRoleController) list(c *gin.Context) {
    c.JSON(200, gin.H{"roles": []any{}})
}
