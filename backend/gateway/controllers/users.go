package controllers

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"gateway/middleware"
	users "gateway/users"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	UsersCodeNotFound      int32 = 1001
	UsersCodeAlreadyExists int32 = 1002
	UsersCodeInvalidInput  int32 = 1003
	UsersCodeInternalError int32 = 2001
)

var usersErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		UsersCodeNotFound:      {http.StatusNotFound, "Ressource introuvable."},
		UsersCodeAlreadyExists: {http.StatusConflict, "Cette ressource existe déjà."},
		UsersCodeInvalidInput:  {http.StatusBadRequest, "Données invalides."},
		UsersCodeInternalError: {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service utilisateurs indisponible.",
}

func UserRoutes(r *gin.RouterGroup) {
	address := os.Getenv("USER_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50052"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to users gRPC server: %v", err)
	}
	client := users.NewUserServiceClient(conn)

	me := r.Group("/me")
	me.GET("", func(c *gin.Context) { GetMe(c, client) })
	me.PUT("", func(c *gin.Context) { UpdateMe(c, client) })
	me.DELETE("", func(c *gin.Context) { DeleteMe(c, client) })

	clients := r.Group("/clients")
	clients.GET("", func(c *gin.Context) { ListClients(c, client) })
	clients.POST("", func(c *gin.Context) { CreateClient(c, client) })
	clients.GET("/:clientId", func(c *gin.Context) { GetClient(c, client) })
	clients.PUT("/:clientId", func(c *gin.Context) { UpdateClient(c, client) })
	clients.DELETE("/:clientId", func(c *gin.Context) { ArchiveClient(c, client) })

	addresses := r.Group("/addresses")
	addresses.GET("", func(c *gin.Context) { ListAddresses(c, client) })
	addresses.POST("", func(c *gin.Context) { CreateAddress(c, client) })
	addresses.GET("/:id", func(c *gin.Context) { GetAddress(c, client) })
	addresses.PUT("/:id", func(c *gin.Context) { UpdateAddress(c, client) })
	addresses.DELETE("/:id", func(c *gin.Context) { ArchiveAddress(c, client) })

	countries := r.Group("/countries")
	countries.GET("", func(c *gin.Context) { ListCountries(c, client) })
	countries.POST("", func(c *gin.Context) { CreateCountry(c, client) })
	countries.GET("/:id", func(c *gin.Context) { GetCountry(c, client) })
	countries.PUT("/:id", func(c *gin.Context) { UpdateCountry(c, client) })
	countries.DELETE("/:id", func(c *gin.Context) { DeleteCountry(c, client) })

	groups := r.Group("/country-groups")
	groups.GET("", func(c *gin.Context) { ListCountryGroups(c, client) })
	groups.POST("", func(c *gin.Context) { CreateCountryGroup(c, client) })
	groups.GET("/:id", func(c *gin.Context) { GetCountryGroup(c, client) })
	groups.PUT("/:id", func(c *gin.Context) { UpdateCountryGroup(c, client) })
	groups.DELETE("/:id", func(c *gin.Context) { DeleteCountryGroup(c, client) })
	groups.POST("/:id/countries/:countryId", func(c *gin.Context) { AttachCountry(c, client) })
	groups.DELETE("/:id/countries/:countryId", func(c *gin.Context) { DetachCountry(c, client) })

	taxes := r.Group("/taxes")
	taxes.GET("", func(c *gin.Context) { ListTaxes(c, client) })
	taxes.GET("/available", func(c *gin.Context) { ListTaxesForUser(c, client) })
	taxes.POST("", func(c *gin.Context) { CreateTax(c, client) })
	taxes.GET("/:id", func(c *gin.Context) { GetTax(c, client) })
	taxes.PUT("/:id", func(c *gin.Context) { UpdateTax(c, client) })
	taxes.DELETE("/:id", func(c *gin.Context) { DeleteTax(c, client) })
}

func userIDFromCtx(c *gin.Context) string {
	id, _ := c.Get(middleware.CtxUserID)
	s, _ := id.(string)
	return s
}

// Rejects javascript:, data:, file:, etc. — defense in depth against
// injecting an unsafe value into the PDF template's <img src>.
func validHTTPURL(s string) bool {
	if s == "" {
		return true
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func paramInt32(c *gin.Context, name string) (int32, bool) {
	v, err := strconv.ParseInt(c.Param(name), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Paramètre invalide."})
		return 0, false
	}
	return int32(v), true
}

func GetMe(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.GetUser(c.Request.Context(), &users.GetUserRequest{UserId: userIDFromCtx(c)})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "user": resp.User})
}

func UpdateMe(c *gin.Context, client users.UserServiceClient) {
	var input struct {
		Phone   string `json:"phone"`
		Company string `json:"company"`
		Siren   string `json:"siren"`
		Vat     string `json:"vat"`
		LogoURL string `json:"logo_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	if !validHTTPURL(input.LogoURL) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "URL du logo invalide."})
		return
	}
	resp, err := client.UpdateUser(c.Request.Context(), &users.UpdateUserRequest{
		UserId:  userIDFromCtx(c),
		Phone:   input.Phone,
		Company: input.Company,
		Siren:   input.Siren,
		Vat:     input.Vat,
		LogoUrl: input.LogoURL,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteMe(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.DeleteUser(c.Request.Context(), &users.DeleteUserRequest{UserId: userIDFromCtx(c)})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type addressInput struct {
	OwnerType        string `json:"owner_type" binding:"required"`
	OwnerID          string `json:"owner_id" binding:"required"`
	Name             string `json:"name" binding:"required"`
	Street           string `json:"street" binding:"required"`
	AdditionalStreet string `json:"additional_street"`
	City             string `json:"city" binding:"required"`
	ZipCode          string `json:"zip_code" binding:"required"`
	CountryID        int32  `json:"country_id" binding:"required"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
}

func parseOwnerType(s string) (users.OwnerType, bool) {
	switch s {
	case "user":
		return users.OwnerType_OWNER_TYPE_USER, true
	case "client":
		return users.OwnerType_OWNER_TYPE_CLIENT, true
	default:
		return users.OwnerType_OWNER_TYPE_UNSPECIFIED, false
	}
}

// ownerTypeToString is the inverse of parseOwnerType for response marshalling.
func ownerTypeToString(t users.OwnerType) string {
	switch t {
	case users.OwnerType_OWNER_TYPE_USER:
		return "user"
	case users.OwnerType_OWNER_TYPE_CLIENT:
		return "client"
	default:
		return ""
	}
}

func ListAddresses(c *gin.Context, client users.UserServiceClient) {
	ownerType, ok := parseOwnerType(c.Query("owner_type"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "owner_type invalide."})
		return
	}
	resp, err := client.ListAddresses(c.Request.Context(), &users.ListAddressesRequest{
		OwnerType:  ownerType,
		OwnerId:    c.Query("owner_id"),
		AuthUserId: userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"addresses": marshalAddresses(resp.Addresses),
	})
}

func GetAddress(c *gin.Context, client users.UserServiceClient) {
	ownerType, ok := parseOwnerType(c.Query("owner_type"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "owner_type invalide."})
		return
	}
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.GetAddress(c.Request.Context(), &users.GetAddressRequest{
		AddressId:  id,
		OwnerType:  ownerType,
		OwnerId:    c.Query("owner_id"),
		AuthUserId: userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"address": marshalAddress(resp.Address),
	})
}

func ArchiveAddress(c *gin.Context, client users.UserServiceClient) {
	ownerType, ok := parseOwnerType(c.Query("owner_type"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "owner_type invalide."})
		return
	}
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.ArchiveAddress(c.Request.Context(), &users.ArchiveAddressRequest{
		AddressId:  id,
		OwnerType:  ownerType,
		OwnerId:    c.Query("owner_id"),
		AuthUserId: userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func CreateAddress(c *gin.Context, client users.UserServiceClient) {
	var input addressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	ownerType, ok := parseOwnerType(input.OwnerType)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "owner_type invalide."})
		return
	}
	resp, err := client.CreateAddress(c.Request.Context(), &users.CreateAddressRequest{
		OwnerType:        ownerType,
		OwnerId:          input.OwnerID,
		Name:             input.Name,
		Street:           input.Street,
		AdditionalStreet: input.AdditionalStreet,
		City:             input.City,
		ZipCode:          input.ZipCode,
		CountryId:        input.CountryID,
		Email:            input.Email,
		Phone:            input.Phone,
		AuthUserId:       userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "address_id": resp.AddressId})
}

func UpdateAddress(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	var input addressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	ownerType, ok := parseOwnerType(input.OwnerType)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "owner_type invalide."})
		return
	}
	resp, err := client.UpdateAddress(c.Request.Context(), &users.UpdateAddressRequest{
		AddressId:        id,
		OwnerType:        ownerType,
		OwnerId:          input.OwnerID,
		Name:             input.Name,
		Street:           input.Street,
		AdditionalStreet: input.AdditionalStreet,
		City:             input.City,
		ZipCode:          input.ZipCode,
		CountryId:        input.CountryID,
		Email:            input.Email,
		Phone:            input.Phone,
		AuthUserId:       userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// marshalAddress flattens a proto Address for JSON responses, mapping the
// OwnerType enum back to its wire-string form.
func marshalAddress(a *users.Address) gin.H {
	if a == nil {
		return nil
	}
	return gin.H{
		"id":                a.Id,
		"owner_type":        ownerTypeToString(a.OwnerType),
		"owner_id":          a.OwnerId,
		"name":              a.Name,
		"street":            a.Street,
		"additional_street": a.AdditionalStreet,
		"city":              a.City,
		"zip_code":          a.ZipCode,
		"country_id":        a.CountryId,
		"email":             a.Email,
		"phone":             a.Phone,
		"archived":          a.Archived,
	}
}

func marshalAddresses(in []*users.Address) []gin.H {
	out := make([]gin.H, 0, len(in))
	for _, a := range in {
		out = append(out, marshalAddress(a))
	}
	return out
}

// ─── Client handlers ─────────────────────────────────────────────────────────

// clientInput is shared by POST and PUT. The PUT handler (UpdateClient) treats
// this as a full-replace payload: optional fields not sent in the JSON arrive
// as "" and clear the corresponding DB column. Callers must send the full set
// — omitting a field will silently null it. See client.Update action for the
// SQL-side contract.
type clientInput struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	Siren     string `json:"siren"`
	Vat       string `json:"vat"`
}

func ListClients(c *gin.Context, client users.UserServiceClient) {
	includeArchived := c.Query("archived") == "true"
	resp, err := client.ListClients(c.Request.Context(), &users.ListClientsRequest{
		UserId:          userIDFromCtx(c),
		IncludeArchived: includeArchived,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "clients": resp.Clients})
}

func CreateClient(c *gin.Context, client users.UserServiceClient) {
	var input clientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateClient(c.Request.Context(), &users.CreateClientRequest{
		UserId:    userIDFromCtx(c),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Company:   input.Company,
		Siren:     input.Siren,
		Vat:       input.Vat,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "client_id": resp.ClientId})
}

func GetClient(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.GetClient(c.Request.Context(), &users.GetClientRequest{
		ClientId: c.Param("clientId"),
		UserId:   userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "client": resp.Client})
}

func UpdateClient(c *gin.Context, client users.UserServiceClient) {
	var input clientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateClient(c.Request.Context(), &users.UpdateClientRequest{
		ClientId:  c.Param("clientId"),
		UserId:    userIDFromCtx(c),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Company:   input.Company,
		Siren:     input.Siren,
		Vat:       input.Vat,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ArchiveClient(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.ArchiveClient(c.Request.Context(), &users.ArchiveClientRequest{
		ClientId: c.Param("clientId"),
		UserId:   userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListCountries(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.ListCountries(c.Request.Context(), &users.ListCountriesRequest{})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "countries": resp.Countries})
}

func CreateCountry(c *gin.Context, client users.UserServiceClient) {
	var input struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateCountry(c.Request.Context(), &users.CreateCountryRequest{
		Code: input.Code,
		Name: input.Name,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "country_id": resp.CountryId})
}

func GetCountry(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.GetCountry(c.Request.Context(), &users.GetCountryRequest{CountryId: id})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "country": resp.Country})
}

func UpdateCountry(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	var input struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateCountry(c.Request.Context(), &users.UpdateCountryRequest{
		CountryId: id,
		Code:      input.Code,
		Name:      input.Name,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteCountry(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.DeleteCountry(c.Request.Context(), &users.DeleteCountryRequest{CountryId: id})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListCountryGroups(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.ListCountryGroups(c.Request.Context(), &users.ListCountryGroupsRequest{})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "country_groups": resp.CountryGroups})
}

func CreateCountryGroup(c *gin.Context, client users.UserServiceClient) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateCountryGroup(c.Request.Context(), &users.CreateCountryGroupRequest{Name: input.Name})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "country_group_id": resp.CountryGroupId})
}

func GetCountryGroup(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.GetCountryGroup(c.Request.Context(), &users.GetCountryGroupRequest{CountryGroupId: id})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "country_group": resp.CountryGroup})
}

func UpdateCountryGroup(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateCountryGroup(c.Request.Context(), &users.UpdateCountryGroupRequest{
		CountryGroupId: id,
		Name:           input.Name,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteCountryGroup(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.DeleteCountryGroup(c.Request.Context(), &users.DeleteCountryGroupRequest{CountryGroupId: id})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func AttachCountry(c *gin.Context, client users.UserServiceClient) {
	groupID, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	countryID, ok := paramInt32(c, "countryId")
	if !ok {
		return
	}
	resp, err := client.AttachCountry(c.Request.Context(), &users.AttachCountryRequest{
		CountryGroupId: groupID,
		CountryId:      countryID,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DetachCountry(c *gin.Context, client users.UserServiceClient) {
	groupID, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	countryID, ok := paramInt32(c, "countryId")
	if !ok {
		return
	}
	resp, err := client.DetachCountry(c.Request.Context(), &users.DetachCountryRequest{
		CountryGroupId: groupID,
		CountryId:      countryID,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListTaxes(c *gin.Context, client users.UserServiceClient) {
	var groupID int32
	if raw := c.Query("country_group_id"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Paramètre invalide."})
			return
		}
		groupID = int32(v)
	}
	resp, err := client.ListTaxes(c.Request.Context(), &users.ListTaxesRequest{CountryGroupId: groupID})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "taxes": resp.Taxes})
}

func ListTaxesForUser(c *gin.Context, client users.UserServiceClient) {
	userID := userIDFromCtx(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Non authentifié."})
		return
	}
	resp, err := client.ListTaxesForUser(c.Request.Context(), &users.ListTaxesForUserRequest{UserId: userID})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "taxes": resp.Taxes})
}

func CreateTax(c *gin.Context, client users.UserServiceClient) {
	var input struct {
		Name           string `json:"name" binding:"required"`
		Rate           string `json:"rate" binding:"required"`
		CountryGroupID int32  `json:"country_group_id" binding:"required"`
		IsDefault      bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateTax(c.Request.Context(), &users.CreateTaxRequest{
		Name:           input.Name,
		Rate:           input.Rate,
		CountryGroupId: input.CountryGroupID,
		IsDefault:      input.IsDefault,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "tax_id": resp.TaxId})
}

func GetTax(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.GetTax(c.Request.Context(), &users.GetTaxRequest{TaxId: id})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "tax": resp.Tax})
}

func UpdateTax(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	var input struct {
		Name      string `json:"name"`
		Rate      string `json:"rate"`
		IsDefault bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateTax(c.Request.Context(), &users.UpdateTaxRequest{
		TaxId:     id,
		Name:      input.Name,
		Rate:      input.Rate,
		IsDefault: input.IsDefault,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteTax(c *gin.Context, client users.UserServiceClient) {
	id, ok := paramInt32(c, "id")
	if !ok {
		return
	}
	resp, err := client.DeleteTax(c.Request.Context(), &users.DeleteTaxRequest{TaxId: id})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
