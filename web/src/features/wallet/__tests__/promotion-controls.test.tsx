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
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { expect, test, vi } from 'vitest'

import { BargainCard } from '../components/bargain-card'
import { SpecialTiersSection } from '../components/special-tiers-section'
import type { SpecialTier } from '../types'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))
const tiers: SpecialTier[] = [
  {
    id: 'available',
    name: 'Available pack',
    price: 5,
    quota: 500000,
    total_limit: 0,
    user_limit: 1,
    enabled: true,
    user_can_purchase: true,
  },
  {
    id: 'sold-out',
    name: 'Sold out pack',
    price: 5,
    quota: 500000,
    total_limit: 2,
    remaining_stock: 0,
    user_limit: 1,
    enabled: true,
    user_can_purchase: false,
  },
]
function TierSelection() {
  const [selected, setSelected] = useState<SpecialTier | null>(null)
  return (
    <SpecialTiersSection
      specialTiers={tiers}
      selectedSpecialTier={selected}
      onSelectTier={setSelected}
    />
  )
}
test('available special tier supports keyboard selection while sold out tier stays disabled', async () => {
  const user = userEvent.setup()
  render(<TierSelection />)
  const choice = screen.getByRole('button', { name: 'Available pack' })
  choice.focus()
  await user.keyboard(' ')
  expect(choice).toHaveAttribute('aria-pressed', 'true')
  expect(screen.getByRole('button', { name: 'Sold out pack' })).toBeDisabled()
})
test('disabled bargain activity hides its purchase controls', () => {
  render(
    <BargainCard
      bargainInfo={{
        enabled: false,
        origin_price: 5,
        cut_amount: 0.5,
        max_slashes: 6,
        min_price: 2,
        quota: 500000,
        user_limit: 1,
        user_can_purchase: true,
        has_active_activity: false,
      }}
      onStartActivity={async () => undefined}
      onBuyBargain={() => undefined}
    />
  )
  expect(screen.queryByRole('button')).not.toBeInTheDocument()
})
