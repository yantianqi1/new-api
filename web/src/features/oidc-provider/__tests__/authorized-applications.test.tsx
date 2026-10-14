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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import type { OIDCGrant } from '../api'
import { OIDCAuthorizedApplicationsCard } from '../authorized-applications-card'

let queryClient: QueryClient
let grants: OIDCGrant[]
beforeEach(() => {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'alice', role: 1 })
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  grants = [
    {
      id: 7,
      client_name: 'Drawing App',
      client_id: 'drawing-public',
      scope: 'openid profile email',
      expires_at: 2000000000,
    },
  ]
  vi.spyOn(api, 'get').mockImplementation(async (url) => {
    expect(url).toBe('/api/user/oidc/grants')
    return { data: { success: true, data: grants } }
  })
})
afterEach(() => {
  cleanup()
  queryClient.clear()
  useAuthStore.getState().auth.reset()
  vi.restoreAllMocks()
})

function renderGrants() {
  return render(
    <QueryClientProvider client={queryClient}>
      <OIDCAuthorizedApplicationsCard />
    </QueryClientProvider>
  )
}

test('shows an empty state when the account has no authorized applications', async () => {
  grants = []
  renderGrants()
  expect(await screen.findByText('No authorized applications')).toBeVisible()
  expect(
    screen.queryByRole('button', { name: 'Modify permissions' })
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Revoke' })
  ).not.toBeInTheDocument()
})

test('reviews the grant and submits only a reduced original scope with openid retained', async () => {
  const patch = vi
    .spyOn(api, 'patch')
    .mockImplementation(async (_url, body) => {
      grants = [{ ...grants[0], scope: (body as { scope: string }).scope }]
      return { data: { success: true, data: grants[0] } }
    })
  renderGrants()
  expect(await screen.findByText('Drawing App')).toBeVisible()
  expect(screen.getByText('drawing-public')).toBeVisible()
  expect(screen.getByText('openid profile email')).toBeVisible()
  expect(screen.getByText('Expires at')).toBeVisible()
  const user = userEvent.setup()
  await user.click(screen.getByRole('button', { name: 'Modify permissions' }))
  const dialog = screen.getByRole('dialog')
  expect(
    within(dialog).getByRole('checkbox', { name: 'openid' })
  ).toHaveAttribute('aria-disabled', 'true')
  await user.click(within(dialog).getByRole('checkbox', { name: 'email' }))
  await user.click(within(dialog).getByRole('button', { name: 'Save changes' }))
  await waitFor(() =>
    expect(patch).toHaveBeenCalledWith(
      '/api/user/oidc/grants/7',
      { scope: 'openid profile' },
      expect.objectContaining({ singleUseAuthorization: true })
    )
  )
  expect(await screen.findByText('openid profile')).toBeVisible()
  await user.click(screen.getByRole('button', { name: 'Modify permissions' }))
  expect(
    screen.queryByRole('checkbox', { name: 'email' })
  ).not.toBeInTheDocument()
})

test('revokes an application only after explicit confirmation and refreshes the list', async () => {
  const remove = vi.spyOn(api, 'delete').mockImplementation(async () => {
    grants = []
    return { data: { success: true, data: null } }
  })
  renderGrants()
  await screen.findByText('Drawing App')
  const user = userEvent.setup()
  await user.click(screen.getByRole('button', { name: 'Revoke' }))
  expect(remove).not.toHaveBeenCalled()
  await user.click(
    within(screen.getByRole('alertdialog')).getByRole('button', {
      name: 'Revoke',
    })
  )
  await waitFor(() =>
    expect(remove).toHaveBeenCalledWith(
      '/api/user/oidc/grants/7',
      expect.objectContaining({ singleUseAuthorization: true })
    )
  )
  expect(await screen.findByText('No authorized applications')).toBeVisible()
})
