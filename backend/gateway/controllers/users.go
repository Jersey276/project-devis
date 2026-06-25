package controllers

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	authpb "gateway/auth"
	"gateway/authz"
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

func usersValidationErrors(errs []*users.ValidationError) []FieldError {
	out := make([]FieldError, len(errs))
	for i, e := range errs {
		out[i] = FieldError{Field: e.Field, Message: e.Message}
	}
	return out
}

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

	admin := r.Group("/admin/accounts")
	admin.Use(middleware.AdminRequired())
	admin.GET("", func(c *gin.Context) { ListAdminAccounts(c, client) })
	admin.PUT("/:userId", func(c *gin.Context) { UpdateAdminAccount(c, client) })
	admin.POST("/:userId/suspend", func(c *gin.Context) { SuspendAdminAccount(c, client) })

	clients := r.Group("/clients")
	clients.GET("", func(c *gin.Context) { ListClients(c, client) })
	clients.POST("", func(c *gin.Context) { CreateClient(c, client) })
	// /me routes must be registered before /:clientId to avoid the wildcard matching "me"
	clients.GET("/me", func(c *gin.Context) { GetMyClient(c, client) })
	clients.PUT("/me", func(c *gin.Context) { UpdateMyClient(c, client) })
	clients.GET("/me/addresses", func(c *gin.Context) { ListMyClientAddresses(c, client) })
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
	countries.Use(middleware.RequireAdminResource(authz.ResourceAdminCountries))
	countries.GET("", func(c *gin.Context) { ListCountries(c, client) })
	countries.POST("", func(c *gin.Context) { CreateCountry(c, client) })
	countries.GET("/:id", func(c *gin.Context) { GetCountry(c, client) })
	countries.PUT("/:id", func(c *gin.Context) { UpdateCountry(c, client) })
	countries.DELETE("/:id", func(c *gin.Context) { DeleteCountry(c, client) })

	groups := r.Group("/country-groups")
	groups.Use(middleware.RequireAdminResource(authz.ResourceAdminCountryGroup))
	groups.GET("", func(c *gin.Context) { ListCountryGroups(c, client) })
	groups.POST("", func(c *gin.Context) { CreateCountryGroup(c, client) })
	groups.GET("/:id", func(c *gin.Context) { GetCountryGroup(c, client) })
	groups.PUT("/:id", func(c *gin.Context) { UpdateCountryGroup(c, client) })
	groups.DELETE("/:id", func(c *gin.Context) { DeleteCountryGroup(c, client) })
	groups.POST("/:id/countries/:countryId", func(c *gin.Context) { AttachCountry(c, client) })
	groups.DELETE("/:id/countries/:countryId", func(c *gin.Context) { DetachCountry(c, client) })

	taxes := r.Group("/taxes")
	taxes.GET("/available", func(c *gin.Context) { ListTaxesForUser(c, client) })

	adminTaxes := taxes.Group("")
	adminTaxes.Use(middleware.RequireAdminResource(authz.ResourceAdminTaxes))
	adminTaxes.GET("", func(c *gin.Context) { ListTaxes(c, client) })
	adminTaxes.POST("", func(c *gin.Context) { CreateTax(c, client) })
	adminTaxes.GET("/:id", func(c *gin.Context) { GetTax(c, client) })
	adminTaxes.PUT("/:id", func(c *gin.Context) { UpdateTax(c, client) })
	adminTaxes.DELETE("/:id", func(c *gin.Context) { DeleteTax(c, client) })
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
	c.JSON(http.StatusOK, gin.H{"success": true, "user": marshalUser(resp.User)})
}

func ensureUserWritable(c *gin.Context, client users.UserServiceClient) bool {
	resp, err := client.GetUserAccessInfo(c.Request.Context(), &users.GetUserAccessInfoRequest{UserId: userIDFromCtx(c)})
	if err != nil {
		usersErrors.unavailable(c)
		return false
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return false
	}
	if resp.Suspended {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "Compte suspendu. Modification désactivée."})
		return false
	}
	return true
}

func UpdateMe(c *gin.Context, client users.UserServiceClient) {
	if !ensureUserWritable(c, client) {
		return
	}
	var input struct {
		Phone      string `json:"phone"`
		Company    string `json:"company"`
		Siren      string `json:"siren"`
		Vat        string `json:"vat"`
		Siret      string `json:"siret"`
		LogoURL    string `json:"logo_url"`
		OssEnabled bool   `json:"oss_enabled"`
		Iban       string `json:"iban"`
		Bic        string `json:"bic"`
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
		UserId:     userIDFromCtx(c),
		Phone:      input.Phone,
		Company:    input.Company,
		Siren:      input.Siren,
		Vat:        input.Vat,
		Siret:      input.Siret,
		LogoUrl:    input.LogoURL,
		OssEnabled: input.OssEnabled,
		Iban:       input.Iban,
		Bic:        input.Bic,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
			return
		}
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteMe(c *gin.Context, client users.UserServiceClient) {
	if !ensureUserWritable(c, client) {
		return
	}
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

func marshalAdminAccount(a *users.AdminAccount) gin.H {
	if a == nil {
		return nil
	}

	var lastLoginAt any = nil
	if a.LastLoginAt != "" {
		lastLoginAt = a.LastLoginAt
	}

	return gin.H{
		"user_id":       a.UserId,
		"first_name":    a.FirstName,
		"last_name":     a.LastName,
		"email":         a.Email,
		"role":          a.Role,
		"plan":          a.Plan,
		"last_login_at": lastLoginAt,
		"suspended":     a.Suspended,
		"phone":         a.Phone,
		"company":       a.Company,
		"siren":         a.Siren,
		"vat":           a.Vat,
	}
}

func ListAdminAccounts(c *gin.Context, client users.UserServiceClient) {
	req := &users.ListAdminAccountsRequest{
		Search: c.Query("search"),
	}
	if roles := c.Query("roles"); roles != "" {
		req.Roles = strings.Split(roles, ",")
	}
	if statuses := c.Query("statuses"); statuses != "" {
		req.Statuses = strings.Split(statuses, ",")
	}

	resp, err := client.ListAdminAccounts(c.Request.Context(), req)
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}

	out := make([]gin.H, 0, len(resp.Accounts))
	for _, account := range resp.Accounts {
		out = append(out, marshalAdminAccount(account))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "users": out})
}

func UpdateAdminAccount(c *gin.Context, client users.UserServiceClient) {
	var input struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email" binding:"required"`
		Role      string `json:"role" binding:"required"`
		Plan      string `json:"plan"`
		Phone     string `json:"phone"`
		Company   string `json:"company"`
		Siren     string `json:"siren"`
		Vat       string `json:"vat"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	resp, err := client.UpdateAdminAccount(c.Request.Context(), &users.UpdateAdminAccountRequest{
		UserId:    c.Param("userId"),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Role:      input.Role,
		Plan:      input.Plan,
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

	authRole := "free_user"
	if input.Role == "admin" {
		authRole = "super_admin"
	}
	if authClient, authClientErr := middleware.GetAuthServiceClient(); authClientErr != nil {
		log.Printf("UpdateAdminAccount: failed to get auth client: %v", authClientErr)
	} else if _, authErr := authClient.UpdateRole(c.Request.Context(), &authpb.UpdateRoleRequest{
		UserId: c.Param("userId"),
		Role:   authRole,
	}); authErr != nil {
		log.Printf("UpdateAdminAccount: failed to update role in auth for user %s: %v", c.Param("userId"), authErr)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func SuspendAdminAccount(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.SuspendAdminAccount(c.Request.Context(), &users.SuspendAdminAccountRequest{UserId: c.Param("userId")})
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Company    string `json:"company"`
	Siren      string `json:"siren"`
	Vat        string `json:"vat"`
	Siret      string `json:"siret"`
	ClientType string `json:"client_type"`
}

// clientTypeFromInput maps the JSON client_type string ("individual"/"business")
// to the proto enum. An empty or unknown value maps to UNSPECIFIED, which the
// users service defaults to "individual" (B2C).
func clientTypeFromInput(s string) users.ClientType {
	switch s {
	case "individual":
		return users.ClientType_CLIENT_TYPE_INDIVIDUAL
	case "business":
		return users.ClientType_CLIENT_TYPE_BUSINESS
	default:
		return users.ClientType_CLIENT_TYPE_UNSPECIFIED
	}
}

// clientTypeToString is the inverse of clientTypeFromInput for response
// marshalling.
func clientTypeToString(t users.ClientType) string {
	switch t {
	case users.ClientType_CLIENT_TYPE_INDIVIDUAL:
		return "individual"
	case users.ClientType_CLIENT_TYPE_BUSINESS:
		return "business"
	default:
		return ""
	}
}

// marshalClient renders a proto Client as a JSON object with the client_type
// enum projected to its string form, mirroring marshalAddress.
func marshalClient(cl *users.Client) gin.H {
	if cl == nil {
		return nil
	}
	return gin.H{
		"client_id":      cl.ClientId,
		"user_id":        cl.UserId,
		"first_name":     cl.FirstName,
		"last_name":      cl.LastName,
		"email":          cl.Email,
		"phone":          cl.Phone,
		"company":        cl.Company,
		"siren":          cl.Siren,
		"vat":            cl.Vat,
		"siret":          cl.Siret,
		"archived":       cl.Archived,
		"client_type":    clientTypeToString(cl.ClientType),
		"linked_user_id": cl.LinkedUserId,
	}
}

func marshalClients(in []*users.Client) []gin.H {
	out := make([]gin.H, 0, len(in))
	for _, cl := range in {
		out = append(out, marshalClient(cl))
	}
	return out
}

// marshalUser renders a proto User as a JSON object. Explicit projection avoids
// the protobuf-go `omitempty` tags dropping false booleans (oss_enabled), which
// the front-end relies on reading explicitly.
func marshalUser(u *users.User) gin.H {
	if u == nil {
		return nil
	}
	return gin.H{
		"user_id":     u.UserId,
		"email":       u.Email,
		"phone":       u.Phone,
		"company":     u.Company,
		"siren":       u.Siren,
		"vat":         u.Vat,
		"siret":       u.Siret,
		"logo_url":    u.LogoUrl,
		"suspended":   u.Suspended,
		"oss_enabled": u.OssEnabled,
		"iban":        u.Iban,
		"bic":         u.Bic,
	}
}

func marshalCountry(co *users.Country) gin.H {
	if co == nil {
		return nil
	}
	return gin.H{
		"id":    co.Id,
		"code":  co.Code,
		"name":  co.Name,
		"is_eu": co.IsEu,
	}
}

func marshalCountries(in []*users.Country) []gin.H {
	out := make([]gin.H, 0, len(in))
	for _, co := range in {
		out = append(out, marshalCountry(co))
	}
	return out
}

func ListClients(c *gin.Context, client users.UserServiceClient) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	var clientTypes []string
	if raw := c.Query("client_types"); raw != "" {
		clientTypes = strings.Split(raw, ",")
	}

	resp, err := client.ListClients(c.Request.Context(), &users.ListClientsRequest{
		UserId:          userIDFromCtx(c),
		IncludeArchived: c.Query("archived") == "true",
		Page:            int32(page),
		PageSize:        int32(pageSize),
		Filters: &users.ClientFilters{
			Search:      c.Query("search"),
			ClientTypes: clientTypes,
		},
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "clients": marshalClients(resp.Clients), "total": resp.Total})
}

func CreateClient(c *gin.Context, client users.UserServiceClient) {
	var input clientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.CreateClient(c.Request.Context(), &users.CreateClientRequest{
		UserId:     userIDFromCtx(c),
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Email:      input.Email,
		Phone:      input.Phone,
		Company:    input.Company,
		Siren:      input.Siren,
		Vat:        input.Vat,
		Siret:      input.Siret,
		ClientType: clientTypeFromInput(input.ClientType),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
	c.JSON(http.StatusOK, gin.H{"success": true, "client": marshalClient(resp.Client)})
}

func UpdateClient(c *gin.Context, client users.UserServiceClient) {
	var input clientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateClient(c.Request.Context(), &users.UpdateClientRequest{
		ClientId:   c.Param("clientId"),
		UserId:     userIDFromCtx(c),
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Email:      input.Email,
		Phone:      input.Phone,
		Company:    input.Company,
		Siren:      input.Siren,
		Vat:        input.Vat,
		Siret:      input.Siret,
		ClientType: clientTypeFromInput(input.ClientType),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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

// ─── Customer /clients/me routes ─────────────────────────────────────────────

// GetMyClients returns all client records linked to the authenticated user (one per provider).
func GetMyClient(c *gin.Context, client users.UserServiceClient) {
	resp, err := client.GetClientsByLinkedUser(c.Request.Context(), &users.GetClientByLinkedUserRequest{
		LinkedUserId: userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, len(resp.Clients))
	for i, cl := range resp.Clients {
		out[i] = marshalClient(cl)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "clients": out})
}

// resolveMyClient finds the linked client for the current user, optionally
// scoped to a specific client_id via query param (required when multiple providers).
func resolveMyClient(c *gin.Context, client users.UserServiceClient) *users.Client {
	resp, err := client.GetClientsByLinkedUser(c.Request.Context(), &users.GetClientByLinkedUserRequest{
		LinkedUserId: userIDFromCtx(c),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return nil
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return nil
	}
	if len(resp.Clients) == 0 {
		usersErrors.reply(c, UsersCodeNotFound)
		return nil
	}

	// If a specific client_id is requested, find it; otherwise use the only one.
	clientID := c.Query("client_id")
	if clientID == "" {
		if len(resp.Clients) == 1 {
			return resp.Clients[0]
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "client_id requis (plusieurs prestataires liés)."})
		return nil
	}
	for _, cl := range resp.Clients {
		if cl.ClientId == clientID {
			return cl
		}
	}
	usersErrors.reply(c, UsersCodeNotFound)
	return nil
}

func UpdateMyClient(c *gin.Context, client users.UserServiceClient) {
	var input clientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	linked := resolveMyClient(c, client)
	if linked == nil {
		return
	}

	resp, err := client.UpdateClient(c.Request.Context(), &users.UpdateClientRequest{
		ClientId:   linked.ClientId,
		UserId:     linked.UserId,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Email:      input.Email,
		Phone:      input.Phone,
		Company:    input.Company,
		Siren:      input.Siren,
		Vat:        input.Vat,
		Siret:      input.Siret,
		ClientType: clientTypeFromInput(input.ClientType),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListMyClientAddresses(c *gin.Context, client users.UserServiceClient) {
	linked := resolveMyClient(c, client)
	if linked == nil {
		return
	}

	resp, err := client.ListAddresses(c.Request.Context(), &users.ListAddressesRequest{
		OwnerType:  users.OwnerType_OWNER_TYPE_CLIENT,
		OwnerId:    linked.ClientId,
		AuthUserId: linked.UserId,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "addresses": marshalAddresses(resp.Addresses)})
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
	c.JSON(http.StatusOK, gin.H{"success": true, "countries": marshalCountries(resp.Countries)})
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
	if status, _ := c.Get(middleware.CtxAccountStatus); status == "suspended" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Compte suspendu: accès restreint.",
			"code":    "ACCOUNT_SUSPENDED",
		})
		return
	}

	userID := userIDFromCtx(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Non authentifié."})
		return
	}

	var includeIDs []int32
	if raw := c.Query("include_ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			v, err := strconv.ParseInt(part, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Paramètre include_ids invalide."})
				return
			}
			includeIDs = append(includeIDs, int32(v))
		}
	}

	var addressID int32
	if raw := c.Query("address_id"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || v <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Paramètre address_id invalide."})
			return
		}
		addressID = int32(v)
	}

	resp, err := client.ListTaxesForUser(c.Request.Context(), &users.ListTaxesForUserRequest{
		UserId:     userID,
		IncludeIds: includeIDs,
		AddressId:  addressID,
	})
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
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
		if len(resp.ValidationErrors) > 0 {
			usersErrors.replyWithValidation(c, resp.Code, usersValidationErrors(resp.ValidationErrors))
		} else {
			usersErrors.reply(c, resp.Code)
		}
		return
	}
	// tax_id may differ from the request id if the update created a new
	// version (the previous row is now superseded). Frontend uses it to
	// refresh local state.
	c.JSON(http.StatusOK, gin.H{"success": true, "tax_id": resp.TaxId})
}

// DeleteTax retires the tax (sets superseded_at). The row is preserved so
// existing quote_lines keep their snapshot. UI surfaces this as "Retirer".
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
