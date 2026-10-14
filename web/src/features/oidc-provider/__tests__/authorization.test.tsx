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
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { OIDCAuthorize } from '../authorize'

beforeEach(() => {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'alice', role: 1 })
  window.history.replaceState(
    {},
    '',
    '/oidc/authorize?client_id=app&consent=allow'
  )
})
afterEach(() => {
  cleanup()
  useAuthStore.getState().auth.reset()
  vi.restoreAllMocks()
  window.history.replaceState({}, '', '/')
})

test('untrusted applications display consent and send only the immutable token and explicit decision', async () => {
  const get = vi.spyOn(api, 'get').mockResolvedValue({
    data: {
      success: true,
      data: {
        client_name: 'Drawing App',
        scope: 'openid profile',
        consent_token: 'session-bound-consent',
        expires_in: 900,
      },
    },
  })
  const post = vi
    .spyOn(api, 'post')
    .mockResolvedValue({ data: { success: false, message: 'Consent expired' } })
  render(<OIDCAuthorize />)
  await screen.findByText('Drawing App')
  expect(screen.getByText(/openid profile/)).toBeInTheDocument()
  expect(screen.getByText(/15/)).toBeInTheDocument()
  expect(post).not.toHaveBeenCalled()
  await userEvent.click(screen.getByRole('button', { name: 'Deny' }))
  expect(get).toHaveBeenCalledWith(
    '/oauth/authorize?client_id=app&consent=allow',
    expect.objectContaining({
      headers: { Accept: 'application/json' },
      singleUseAuthorization: true,
    })
  )
  expect(post).toHaveBeenCalledWith(
    '/oauth/authorize',
    { consent_token: 'session-bound-consent', decision: 'deny' },
    expect.objectContaining({ singleUseAuthorization: true })
  )
  await screen.findByText('Authorization failed')
  expect(
    screen.queryByRole('button', { name: 'Authorize' })
  ).not.toBeInTheDocument()
})

test('invalid authorization destinations produce an error without executing a browser redirect', async () => {
  vi.spyOn(api, 'get').mockResolvedValue({
    data: { success: true, data: { authorization_url: 'javascript:alert(1)' } },
  })
  render(<OIDCAuthorize />)
  await screen.findByText('Authorization failed')
  expect(screen.getByText('Invalid authorization redirect')).toBeInTheDocument()
})
