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

import { WalletStatsCard } from '../components/wallet-stats-card'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}))
test('wallet exposes paid and bonus balances as separate accounts', () => {
  render(
    <WalletStatsCard
      user={{
        id: 1,
        username: 'alice',
        quota: 500000,
        bonus_quota: 100000,
        used_quota: 0,
        request_count: 0,
        aff_quota: 0,
        aff_history_quota: 0,
        aff_count: 0,
        group: 'default',
      }}
    />
  )
  expect(screen.getByText('Paid Balance')).toBeInTheDocument()
  expect(screen.getByText('Points Balance')).toBeInTheDocument()
})
