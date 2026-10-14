/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { api } from '@/lib/api'
import { authRequestOptions, authResult } from '@/lib/secure-verification'

export interface OIDCClient {
  id: number
  name: string
  client_id: string
  redirect_uris: string
  scopes: string
  trusted: boolean
  enabled: boolean
}
export interface OIDCClientInput {
  name: string
  redirect_uris: string[]
  scopes: string[]
  trusted: boolean
  enabled: boolean
}
export function listOIDCClients(): Promise<OIDCClient[]> {
  return authResult(api.get('/api/oidc/client/', authRequestOptions))
}
export async function mutateOIDCClient(
  action: 'create' | 'update' | 'delete' | 'secret.rotate',
  id: number | undefined,
  input: OIDCClientInput | undefined,
  proof: string,
  signal: AbortSignal
): Promise<(OIDCClient & { client_secret?: string }) | null> {
  const config = {
    ...authRequestOptions,
    headers: { 'X-Security-Proof': proof },
    singleUseAuthorization: true,
    signal,
  }
  if (action === 'create') {
    return authResult(api.post('/api/oidc/client/', input, config))
  }
  if (action === 'update') {
    return authResult(api.put(`/api/oidc/client/${id}`, input, config))
  }
  if (action === 'delete') {
    return authResult(api.delete(`/api/oidc/client/${id}`, config))
  }
  return authResult(api.post(`/api/oidc/client/${id}/secret`, {}, config))
}

export interface OIDCGrant {
  id: number
  client_name: string
  client_id: string
  scope: string
  expires_at: number
}

export function listOIDCGrants(signal: AbortSignal): Promise<OIDCGrant[]> {
  return authResult(
    api.get('/api/user/oidc/grants', { ...authRequestOptions, signal })
  )
}

export function reduceOIDCGrantScope(
  id: number,
  scope: string
): Promise<OIDCGrant | null> {
  return authResult(
    api.patch(
      `/api/user/oidc/grants/${id}`,
      { scope },
      { ...authRequestOptions, singleUseAuthorization: true }
    )
  )
}

export function revokeOIDCGrant(id: number): Promise<null> {
  return authResult(
    api.delete(`/api/user/oidc/grants/${id}`, {
      ...authRequestOptions,
      singleUseAuthorization: true,
    })
  )
}
