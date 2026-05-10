import { apiFetch, type ApiResult } from "@/lib/api";

export type AddressOwner =
  | { type: "user"; userId: string }
  | { type: "client"; clientId: string };

// buildOwner converts the scalar (ownerType, ownerId) form used by table /
// drawer props into the discriminated-union AddressOwner the service layer
// expects. Centralised here so the components don't drift in shape.
export function buildOwner(
  ownerType: "user" | "client",
  ownerId: string,
): AddressOwner {
  return ownerType === "user"
    ? { type: "user", userId: ownerId }
    : { type: "client", clientId: ownerId };
}

export type AddressPayload = {
  name: string;
  street: string;
  additional_street?: string;
  city: string;
  zip_code: string;
  country_id: number | null;
  email?: string;
  phone?: string;
};

const BASE = "/api/users/addresses";

function ownerId(owner: AddressOwner): string {
  switch (owner.type) {
    case "user":
      return owner.userId;
    case "client":
      return owner.clientId;
    default: {
      const _exhaustive: never = owner;
      return _exhaustive;
    }
  }
}

function ownerQuery(owner: AddressOwner): string {
  return new URLSearchParams({
    owner_type: owner.type,
    owner_id: ownerId(owner),
  }).toString();
}

function ownerBody(owner: AddressOwner) {
  return { owner_type: owner.type, owner_id: ownerId(owner) };
}

export async function listAddresses(owner: AddressOwner): Promise<ApiResult> {
  return apiFetch(`${BASE}?${ownerQuery(owner)}`);
}

export async function getAddress(
  owner: AddressOwner,
  id: number,
): Promise<ApiResult> {
  return apiFetch(`${BASE}/${id}?${ownerQuery(owner)}`);
}

export async function createAddress(
  owner: AddressOwner,
  payload: AddressPayload,
): Promise<ApiResult> {
  return apiFetch(BASE, {
    method: "POST",
    body: JSON.stringify({ ...ownerBody(owner), ...payload }),
  });
}

export async function updateAddress(
  owner: AddressOwner,
  id: number,
  payload: AddressPayload,
): Promise<ApiResult> {
  return apiFetch(`${BASE}/${id}`, {
    method: "PUT",
    body: JSON.stringify({ ...ownerBody(owner), ...payload }),
  });
}

export async function archiveAddress(
  owner: AddressOwner,
  id: number,
): Promise<ApiResult> {
  return apiFetch(`${BASE}/${id}?${ownerQuery(owner)}`, {
    method: "DELETE",
  });
}
