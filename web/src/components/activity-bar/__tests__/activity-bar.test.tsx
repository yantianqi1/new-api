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
import { afterEach, expect, test, vi } from 'vitest'

import { ActivityBar } from '../index'

const status = vi.hoisted(() => ({ data: {} as Record<string, unknown> }))
vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({ status: status.data }),
}))
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))
vi.mock('@tanstack/react-router', () => ({
  Link: (props: { to: string; children: React.ReactNode }) => (
    <a href={props.to}>{props.children}</a>
  ),
}))
afterEach(() => {
  cleanup()
  status.data = {}
})

test('configured activities display safe internal and external destinations', () => {
  status.data = {
    activity_bar_enabled: true,
    activity_bar_items: [
      { title: 'Recharge', link: '/wallet', tag: 'New' },
      { title: 'Rewards', link: 'https://gift.inkwan.app' },
      { title: 'Unsafe', link: 'javascript:alert(1)' },
    ],
  }
  render(<ActivityBar />)
  expect(screen.getByRole('link', { name: 'Recharge New' })).toHaveAttribute(
    'href',
    '/wallet'
  )
  expect(screen.getByRole('link', { name: 'Rewards' })).toHaveAttribute(
    'rel',
    'noopener noreferrer'
  )
  expect(screen.queryByRole('link', { name: 'Unsafe' })).not.toBeInTheDocument()
})

test('disabled activities and invalid configurations render no marquee', () => {
  status.data = {
    activity_bar_enabled: false,
    activity_bar_items: [{ title: 'Recharge', link: '/wallet' }],
  }
  const view = render(<ActivityBar />)
  expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  status.data = {
    activity_bar_enabled: true,
    activity_bar_items: 'invalid json',
  }
  view.rerender(<ActivityBar />)
  expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
})
