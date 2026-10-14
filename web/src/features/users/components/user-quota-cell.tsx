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

import { QuotaDetailsPopover } from '@/components/quota-details-popover'
import { StatusBadge } from '@/components/status-badge'
import { formatQuotaWithCurrency, getCurrencyDisplay } from '@/lib/currency'
import { cn } from '@/lib/utils'
import { useSystemConfigStore } from '@/stores/system-config-store'

type UserQuotaCellProps = {
  remaining: number
  bonusQuota?: number
  used: number
}

export function UserQuotaCell(props: UserQuotaCellProps) {
  const { t } = useTranslation()
  useSystemConfigStore((state) => state.config.currency)

  const { meta: currency } = getCurrencyDisplay()
  const quotaUnit = currency.kind === 'tokens' ? t('Tokens') : currency.symbol
  const bonusQuota = props.bonusQuota ?? 0
  const totalRemaining = props.remaining + bonusQuota
  const hasQuota = props.remaining !== 0 || bonusQuota !== 0 || props.used !== 0
  const formattedRemaining = formatQuotaWithCurrency(props.remaining, {
    showSymbol: false,
  })
  const formattedBonus = formatQuotaWithCurrency(bonusQuota, {
    showSymbol: false,
  })
  const formattedTotalRemaining = formatQuotaWithCurrency(totalRemaining, {
    showSymbol: false,
  })
  const formattedUsed = formatQuotaWithCurrency(props.used, {
    showSymbol: false,
  })

  const details = [
    { label: t('Available Balance'), value: formattedTotalRemaining },
  ]
  if (bonusQuota > 0) {
    details.push({ label: t('Paid Balance'), value: formattedRemaining })
    details.push({ label: t('Points Balance'), value: formattedBonus })
  }
  details.push({ label: t('Total Used'), value: formattedUsed })

  return (
    <QuotaDetailsPopover
      title={`${t('Quota')} (${quotaUnit})`}
      triggerLabel={
        hasQuota
          ? `${t('Available Balance')} ${formattedTotalRemaining}; ${t('Used amount')} ${formattedUsed}`
          : t('No Quota')
      }
      details={details}
    >
      {hasQuota ? (
        <span className='grid min-w-0 grid-cols-1 gap-y-1 text-sm tabular-nums'>
          <span
            className={cn(
              totalRemaining < 0 && 'text-destructive',
              totalRemaining === 0 && 'text-muted-foreground'
            )}
          >
            {formattedTotalRemaining}
            {bonusQuota > 0 && (
              <span className='text-muted-foreground ml-1 text-xs font-normal'>
                ({formattedRemaining} + {formattedBonus})
              </span>
            )}
          </span>
          <span
            data-table-text='secondary'
            className='text-muted-foreground flex items-baseline gap-1 text-xs font-normal'
          >
            <span>{t('Used amount')}</span>
            <span>{formattedUsed}</span>
          </span>
        </span>
      ) : (
        <StatusBadge
          label={t('No Quota')}
          variant='neutral'
          copyable={false}
          className='-ml-1.5 font-normal'
        />
      )}
    </QuotaDetailsPopover>
  )
}
