package actions

import (
	"context"

	"project-devis-users/actions/address"
	"project-devis-users/actions/client"
	"project-devis-users/actions/consent"
	"project-devis-users/actions/country"
	country_group "project-devis-users/actions/country_group"
	"project-devis-users/actions/tax"
	"project-devis-users/actions/user"
	usersGrpc "project-devis-users/services/grpc"
)

// ─── User ────────────────────────────────────────────────────────────────────

func (s *Server) CreateUser(ctx context.Context, req *usersGrpc.CreateUserRequest) (*usersGrpc.CreateUserResponse, error) {
	return user.Create(ctx, s.db, req)
}

func (s *Server) GetUser(ctx context.Context, req *usersGrpc.GetUserRequest) (*usersGrpc.GetUserResponse, error) {
	return user.Get(ctx, s.db, req)
}

func (s *Server) UpdateUser(ctx context.Context, req *usersGrpc.UpdateUserRequest) (*usersGrpc.UpdateUserResponse, error) {
	return user.Update(ctx, s.db, req)
}

func (s *Server) DeleteUser(ctx context.Context, req *usersGrpc.DeleteUserRequest) (*usersGrpc.GenericResponse, error) {
	return user.Delete(ctx, s.db, req)
}

func (s *Server) GetUserAccessInfo(ctx context.Context, req *usersGrpc.GetUserAccessInfoRequest) (*usersGrpc.GetUserAccessInfoResponse, error) {
	return user.GetUserAccessInfo(ctx, s.db, req)
}

func (s *Server) GetUserAccessInfoByEmail(ctx context.Context, req *usersGrpc.GetUserAccessInfoByEmailRequest) (*usersGrpc.GetUserAccessInfoResponse, error) {
	return user.GetUserAccessInfoByEmail(ctx, s.db, req)
}

func (s *Server) ListAdminAccounts(ctx context.Context, req *usersGrpc.ListAdminAccountsRequest) (*usersGrpc.ListAdminAccountsResponse, error) {
	return user.ListAdminAccounts(ctx, s.db, req)
}

func (s *Server) UpdateAdminAccount(ctx context.Context, req *usersGrpc.UpdateAdminAccountRequest) (*usersGrpc.GenericResponse, error) {
	return user.UpdateAdminAccount(ctx, s.db, req)
}

func (s *Server) SuspendAdminAccount(ctx context.Context, req *usersGrpc.SuspendAdminAccountRequest) (*usersGrpc.GenericResponse, error) {
	return user.SuspendAdminAccount(ctx, s.db, req)
}

func (s *Server) TouchUserLastLogin(ctx context.Context, req *usersGrpc.TouchUserLastLoginRequest) (*usersGrpc.GenericResponse, error) {
	return user.TouchUserLastLogin(ctx, s.db, req)
}

func (s *Server) UpdateUserEmail(ctx context.Context, req *usersGrpc.UpdateUserEmailRequest) (*usersGrpc.GenericResponse, error) {
	return user.UpdateEmail(ctx, s.db, req)
}

// ─── Client ──────────────────────────────────────────────────────────────────

func (s *Server) CreateClient(ctx context.Context, req *usersGrpc.CreateClientRequest) (*usersGrpc.CreateClientResponse, error) {
	return client.Create(ctx, s.db, req)
}

func (s *Server) GetClient(ctx context.Context, req *usersGrpc.GetClientRequest) (*usersGrpc.GetClientResponse, error) {
	return client.Get(ctx, s.db, req)
}

func (s *Server) ListClients(ctx context.Context, req *usersGrpc.ListClientsRequest) (*usersGrpc.ListClientsResponse, error) {
	return client.List(ctx, s.db, req)
}

func (s *Server) UpdateClient(ctx context.Context, req *usersGrpc.UpdateClientRequest) (*usersGrpc.UpdateClientResponse, error) {
	return client.Update(ctx, s.db, req)
}

func (s *Server) ArchiveClient(ctx context.Context, req *usersGrpc.ArchiveClientRequest) (*usersGrpc.GenericResponse, error) {
	return client.Archive(ctx, s.db, req)
}

func (s *Server) LinkClientUser(ctx context.Context, req *usersGrpc.LinkClientUserRequest) (*usersGrpc.GenericResponse, error) {
	return client.LinkUser(ctx, s.db, req)
}

func (s *Server) GetClientsByLinkedUser(ctx context.Context, req *usersGrpc.GetClientByLinkedUserRequest) (*usersGrpc.ListClientsResponse, error) {
	return client.GetByLinkedUser(ctx, s.db, req)
}

// ─── Address ─────────────────────────────────────────────────────────────────

func (s *Server) CreateAddress(ctx context.Context, req *usersGrpc.CreateAddressRequest) (*usersGrpc.CreateAddressResponse, error) {
	return address.Create(ctx, s.db, req)
}

func (s *Server) GetAddress(ctx context.Context, req *usersGrpc.GetAddressRequest) (*usersGrpc.GetAddressResponse, error) {
	return address.Get(ctx, s.db, req)
}

func (s *Server) ListAddresses(ctx context.Context, req *usersGrpc.ListAddressesRequest) (*usersGrpc.ListAddressesResponse, error) {
	return address.List(ctx, s.db, req)
}

func (s *Server) UpdateAddress(ctx context.Context, req *usersGrpc.UpdateAddressRequest) (*usersGrpc.UpdateAddressResponse, error) {
	return address.Update(ctx, s.db, req)
}

func (s *Server) ArchiveAddress(ctx context.Context, req *usersGrpc.ArchiveAddressRequest) (*usersGrpc.GenericResponse, error) {
	return address.Archive(ctx, s.db, req)
}

// ─── Country ──────────────────────────────────────────────────────────────────

func (s *Server) CreateCountry(ctx context.Context, req *usersGrpc.CreateCountryRequest) (*usersGrpc.CreateCountryResponse, error) {
	return country.Create(ctx, s.db, req)
}

func (s *Server) GetCountry(ctx context.Context, req *usersGrpc.GetCountryRequest) (*usersGrpc.GetCountryResponse, error) {
	return country.Get(ctx, s.db, req)
}

func (s *Server) ListCountries(ctx context.Context, req *usersGrpc.ListCountriesRequest) (*usersGrpc.ListCountriesResponse, error) {
	return country.List(ctx, s.db, req)
}

func (s *Server) UpdateCountry(ctx context.Context, req *usersGrpc.UpdateCountryRequest) (*usersGrpc.UpdateCountryResponse, error) {
	return country.Update(ctx, s.db, req)
}

func (s *Server) DeleteCountry(ctx context.Context, req *usersGrpc.DeleteCountryRequest) (*usersGrpc.GenericResponse, error) {
	return country.Delete(ctx, s.db, req)
}

// ─── CountryGroup ─────────────────────────────────────────────────────────────

func (s *Server) CreateCountryGroup(ctx context.Context, req *usersGrpc.CreateCountryGroupRequest) (*usersGrpc.CreateCountryGroupResponse, error) {
	return country_group.Create(ctx, s.db, req)
}

func (s *Server) GetCountryGroup(ctx context.Context, req *usersGrpc.GetCountryGroupRequest) (*usersGrpc.GetCountryGroupResponse, error) {
	return country_group.Get(ctx, s.db, req)
}

func (s *Server) ListCountryGroups(ctx context.Context, req *usersGrpc.ListCountryGroupsRequest) (*usersGrpc.ListCountryGroupsResponse, error) {
	return country_group.List(ctx, s.db, req)
}

func (s *Server) UpdateCountryGroup(ctx context.Context, req *usersGrpc.UpdateCountryGroupRequest) (*usersGrpc.UpdateCountryGroupResponse, error) {
	return country_group.Update(ctx, s.db, req)
}

func (s *Server) DeleteCountryGroup(ctx context.Context, req *usersGrpc.DeleteCountryGroupRequest) (*usersGrpc.GenericResponse, error) {
	return country_group.Delete(ctx, s.db, req)
}

func (s *Server) AttachCountry(ctx context.Context, req *usersGrpc.AttachCountryRequest) (*usersGrpc.GenericResponse, error) {
	return country_group.AttachCountry(ctx, s.db, req)
}

func (s *Server) DetachCountry(ctx context.Context, req *usersGrpc.DetachCountryRequest) (*usersGrpc.GenericResponse, error) {
	return country_group.DetachCountry(ctx, s.db, req)
}

// ─── Tax ─────────────────────────────────────────────────────────────────────

func (s *Server) CreateTax(ctx context.Context, req *usersGrpc.CreateTaxRequest) (*usersGrpc.CreateTaxResponse, error) {
	return tax.Create(ctx, s.db, req)
}

func (s *Server) GetTax(ctx context.Context, req *usersGrpc.GetTaxRequest) (*usersGrpc.GetTaxResponse, error) {
	return tax.Get(ctx, s.db, req)
}

func (s *Server) ListTaxes(ctx context.Context, req *usersGrpc.ListTaxesRequest) (*usersGrpc.ListTaxesResponse, error) {
	return tax.List(ctx, s.db, req)
}

func (s *Server) ListTaxesForUser(ctx context.Context, req *usersGrpc.ListTaxesForUserRequest) (*usersGrpc.ListTaxesResponse, error) {
	return tax.ListForUser(ctx, s.db, req)
}

func (s *Server) ListTaxesForCountry(ctx context.Context, req *usersGrpc.ListTaxesForCountryRequest) (*usersGrpc.ListTaxesResponse, error) {
	return tax.ListForCountry(ctx, s.db, req)
}

func (s *Server) UpdateTax(ctx context.Context, req *usersGrpc.UpdateTaxRequest) (*usersGrpc.UpdateTaxResponse, error) {
	return tax.Update(ctx, s.db, req)
}

func (s *Server) DeleteTax(ctx context.Context, req *usersGrpc.DeleteTaxRequest) (*usersGrpc.GenericResponse, error) {
	return tax.Delete(ctx, s.db, req)
}

// ─── Consent ─────────────────────────────────────────────────────────────────

func (s *Server) AcceptConsent(ctx context.Context, req *usersGrpc.AcceptConsentRequest) (*usersGrpc.GenericResponse, error) {
	return consent.Accept(ctx, s.db, req)
}

func (s *Server) GetConsentStatus(ctx context.Context, req *usersGrpc.GetConsentStatusRequest) (*usersGrpc.GetConsentStatusResponse, error) {
	return consent.GetStatus(ctx, s.db, req)
}
