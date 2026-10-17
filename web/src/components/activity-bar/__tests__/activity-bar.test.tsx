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
  Link: ({ to, ...props }: { to: string } & React.ComponentProps<'a'>) => (
    <a href={to} {...props} />
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

test.each(['array', 'json'])(
  'disabled items stay hidden in %s configuration',
  (format) => {
    const items = [
      { title: 'Disabled promotion', link: '/wallet', enabled: false },
      { title: 'Current promotion', link: '/pricing', enabled: true },
      { title: 'Legacy promotion', link: '/about' },
    ]
    status.data = {
      activity_bar_enabled: true,
      activity_bar_items: format === 'json' ? JSON.stringify(items) : items,
    }
    const view = render(<ActivityBar />)
    expect(screen.queryAllByText('Disabled promotion')).toHaveLength(0)
    expect(
      screen.getByRole('link', { name: 'Current promotion' })
    ).toHaveAttribute('href', '/pricing')
    expect(
      screen.getByRole('link', { name: 'Legacy promotion' })
    ).toHaveAttribute('href', '/about')

    status.data.activity_bar_items =
      format === 'json' ? JSON.stringify([items[0]]) : [items[0]]
    view.rerender(<ActivityBar />)
    expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  }
)

test('both scrolling copies retain destinations without duplicate keyboard stops', async () => {
  const user = userEvent.setup()
  status.data = {
    activity_bar_enabled: true,
    activity_bar_items: [
      { title: 'Recharge', link: '/wallet', tag: 'New' },
      { title: 'Rewards', link: 'https://gift.inkwan.app' },
      { title: 'Notice' },
      { title: 'Unsafe', link: 'javascript:alert(1)' },
    ],
  }
  render(
    <>
      <ActivityBar />
      <button type='button'>After activities</button>
    </>
  )
  const links = screen.getAllByRole('link', { hidden: true })
  expect(links).toHaveLength(4)
  for (const link of links.filter((item) =>
    item.closest('[aria-hidden="true"]')
  )) {
    expect(link).toHaveAttribute('tabindex', '-1')
  }
  expect(links.map((link) => link.getAttribute('href'))).toEqual([
    '/wallet',
    'https://gift.inkwan.app',
    '/wallet',
    'https://gift.inkwan.app',
  ])
  for (const link of links.filter((item) =>
    item.getAttribute('href')?.startsWith('https:')
  )) {
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  }
  expect(screen.queryByText('Unsafe')).not.toBeInTheDocument()
  await user.tab()
  expect(screen.getByRole('link', { name: 'Recharge New' })).toHaveFocus()
  await user.tab()
  expect(screen.getByRole('link', { name: 'Rewards' })).toHaveFocus()
  await user.tab()
  expect(screen.getByRole('button', { name: 'After activities' })).toHaveFocus()
})
