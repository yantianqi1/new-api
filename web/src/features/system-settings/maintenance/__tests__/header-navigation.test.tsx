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
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { expect, test, vi } from 'vitest'

import { api } from '@/lib/api'

import { SettingsPageProvider } from '../../components/settings-page-context'
import { parseHeaderNavModules, serializeHeaderNavModules } from '../config'
import { HeaderNavigationSection } from '../header-navigation-section'

const saved = JSON.stringify({
  home: true,
  console: true,
  pricing: {
    enabled: true,
    requireAuth: false,
    link: 'https://gift.inkwan.app/models',
  },
  rankings: { enabled: true, requireAuth: true, link: '/rankings-custom' },
  docs: true,
  about: true,
})

function Fixture() {
  const [container, setContainer] = useState<HTMLDivElement | null>(null)
  const config = parseHeaderNavModules(saved)
  return (
    <>
      <div ref={setContainer} />
      <SettingsPageProvider actionsContainer={container}>
        <HeaderNavigationSection
          config={config}
          initialSerialized={serializeHeaderNavModules(config)}
        />
      </SettingsPageProvider>
    </>
  )
}

test('preserves custom navigation destinations while parsing and editing pricing links', async () => {
  const config = parseHeaderNavModules(saved)
  expect(config.pricing.link).toBe('https://gift.inkwan.app/models')
  expect(config.rankings.link).toBe('/rankings-custom')
  const put = vi
    .spyOn(api, 'put')
    .mockResolvedValue({ data: { success: true } })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  render(
    <QueryClientProvider client={client}>
      <Fixture />
    </QueryClientProvider>
  )
  const input = screen.getByRole('textbox', { name: 'Custom Redirect Link' })
  expect(input).toHaveValue('https://gift.inkwan.app/models')
  const user = userEvent.setup()
  await user.clear(input)
  await user.type(input, 'https://gift.inkwan.app/tiers')
  await user.click(screen.getByRole('button', { name: 'Save navigation' }))
  await waitFor(() =>
    expect(put).toHaveBeenCalledWith('/api/option/', {
      key: 'HeaderNavModules',
      value: expect.any(String),
    })
  )
  const payload = put.mock.calls[0][1] as { value: string }
  expect(JSON.parse(payload.value).pricing.link).toBe(
    'https://gift.inkwan.app/tiers'
  )
  expect(JSON.parse(payload.value).rankings.link).toBe('/rankings-custom')
  client.clear()
})
