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
import { render, screen } from '@testing-library/react'
import { expect, test, vi } from 'vitest'

import { Home } from '../index'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, values?: Record<string, string>) =>
      values ? key.replace('{{usageUrl}}', values.usageUrl ?? '') : key,
    i18n: { language: 'en' },
  }),
}))
vi.mock('@/hooks/use-status', () => ({ useStatus: () => ({ status: {} }) }))
vi.mock('../hooks', () => ({
  useHomePageContent: () => ({ content: '', isLoaded: true, isUrl: false }),
}))
vi.mock('@/context/theme-provider', () => ({
  useTheme: () => ({ resolvedTheme: 'light' }),
}))
vi.mock('@/components/layout', () => ({
  PublicLayout: (props: { children: React.ReactNode }) => (
    <main>{props.children}</main>
  ),
}))
vi.mock('@/components/layout/components/footer', () => ({ Footer: () => null }))
vi.mock('@tanstack/react-router', () => ({
  Link: (props: React.ComponentProps<'a'> & { to: string }) => (
    <a href={props.to}>{props.children}</a>
  ),
}))
test('default home retains the station API endpoint copy control', () => {
  render(<Home />)
  expect(
    screen.getByRole('button', { name: 'Copy Base URL' })
  ).toBeInTheDocument()
})

test('the default integration prompt uses English source text and the configured endpoint', () => {
  render(<Home />)
  const prompt = screen.getByText(
    /Help me integrate an API key quota lookup component/
  )
  expect(prompt).toHaveTextContent('Authorization: Bearer <API_KEY>')
  expect(prompt).toHaveTextContent('/api/usage/token/')
  expect(prompt).toHaveTextContent('total_available (remaining quota)')
  expect(prompt).not.toHaveTextContent('请帮我')
})
