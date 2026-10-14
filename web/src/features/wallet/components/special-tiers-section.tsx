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
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { formatQuota } from '@/lib/format'

import type { SpecialTier } from '../types'

interface SpecialTiersSectionProps {
  specialTiers: SpecialTier[]
  selectedSpecialTier: SpecialTier | null
  onSelectTier: (tier: SpecialTier | null) => void
}

export function SpecialTiersSection(props: SpecialTiersSectionProps) {
  const { t } = useTranslation()
  if (props.specialTiers.length === 0) return null
  return (
    <section
      className='flex flex-col gap-3'
      aria-label={t('Limited-Time Special Offers')}
    >
      <h3 className='text-sm font-semibold'>
        {t('Limited-Time Special Offers')}
      </h3>
      <ToggleGroup
        variant='outline'
        className='grid w-full gap-3 sm:grid-cols-2'
        value={props.selectedSpecialTier ? [props.selectedSpecialTier.id] : []}
        onValueChange={(values) =>
          props.onSelectTier(
            props.specialTiers.find((tier) => tier.id === values[0]) ?? null
          )
        }
      >
        {props.specialTiers.map((tier) => {
          const purchased =
            tier.user_purchased_count ?? tier.user_purchased ?? 0
          const reachedLimit =
            tier.user_limit > 0 && purchased >= tier.user_limit
          const soldOut =
            (tier.has_stock_limit || tier.total_limit > 0) &&
            tier.remaining_stock === 0
          const unavailable =
            !tier.enabled ||
            tier.user_can_purchase === false ||
            reachedLimit ||
            soldOut
          let status = t('Special Offer')
          if (reachedLimit) status = t('Already Participated')
          else if (soldOut) status = t('Sold Out')
          return (
            <ToggleGroupItem
              key={tier.id}
              value={tier.id}
              disabled={unavailable}
              aria-label={tier.name}
              className='h-auto min-w-0 flex-col items-start gap-2 rounded-lg p-4 whitespace-normal'
            >
              <span className='flex w-full items-start justify-between gap-2'>
                <span className='break-words'>{tier.name}</span>
                <Badge variant='secondary'>{status}</Badge>
              </span>
              <span>{`¥${tier.price.toFixed(2)}`}</span>
              <span className='text-muted-foreground'>
                {t('Pack Quota')}: {formatQuota(tier.quota)}
              </span>
              {tier.description && (
                <span className='text-muted-foreground text-left text-xs break-words'>
                  {tier.description}
                </span>
              )}
              {tier.user_limit > 0 && (
                <span className='text-xs'>
                  {t('Limit {{limit}} per user', { limit: tier.user_limit })}
                </span>
              )}
              {(tier.is_stock_visible ?? tier.show_stock ?? true) &&
                tier.remaining_stock !== undefined &&
                tier.total_limit > 0 && (
                  <span className='text-xs'>
                    {t('{{stock}} left', { stock: tier.remaining_stock })}
                  </span>
                )}
            </ToggleGroupItem>
          )
        })}
      </ToggleGroup>
    </section>
  )
}
