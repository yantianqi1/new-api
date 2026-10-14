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

import { CopyButton } from '@/components/copy-button'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { formatQuota } from '@/lib/format'

import type { BargainInfo, BargainActivity } from '../types'

interface BargainCardProps {
  bargainInfo?: BargainInfo | null
  onStartActivity: () => Promise<void>
  onBuyBargain: (activity: BargainActivity) => void
  onRefresh?: () => void
  affLink?: string
  startLoading?: boolean
}

export function BargainCard(props: BargainCardProps) {
  const { t } = useTranslation()
  const info = props.bargainInfo
  if (!info?.enabled) return null
  const activity = info.activity
  const progress =
    info.max_slashes > 0
      ? Math.min(100, ((activity?.cut_count ?? 0) / info.max_slashes) * 100)
      : 0
  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {info.title || t('Friend Slash Deal · Value Gift Pack')}
        </CardTitle>
        <CardDescription>
          {info.description ||
            t(
              'Invite friends to register and slash the price, getting quota at the lowest price!'
            )}
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-4'>
        <div className='flex flex-wrap items-center gap-3'>
          <Badge variant='secondary'>
            {t('Pack Quota')}: {formatQuota(activity?.quota ?? info.quota)}
          </Badge>
          {info.user_limit > 0 && (
            <Badge variant='outline'>
              {t('Limit {{limit}} per user', { limit: info.user_limit })}
            </Badge>
          )}
        </div>
        {!info.user_can_purchase && (
          <p className='text-muted-foreground'>{t('Already Participated')}</p>
        )}
        {info.user_can_purchase && !activity && (
          <Button disabled={props.startLoading} onClick={props.onStartActivity}>
            {t('Start Slashing Now')}
          </Button>
        )}
        {info.user_can_purchase && activity && (
          <>
            <div className='flex flex-wrap items-baseline gap-3'>
              <span className='text-muted-foreground line-through'>
                ¥{info.origin_price.toFixed(2)}
              </span>
              <span className='text-2xl font-semibold'>
                ¥{activity.current_price.toFixed(2)}
              </span>
              {activity.current_price <= info.min_price && (
                <Badge>{t('Reached Lowest Price!')}</Badge>
              )}
            </div>
            <p className='text-muted-foreground text-sm'>
              {t('Floor Price')}: ¥{info.min_price.toFixed(2)} ·{' '}
              {t('Per Slash')}: ¥{info.cut_amount.toFixed(2)}
            </p>
            <Progress value={progress} aria-label={t('Slash Progress')} />
            <p className='text-muted-foreground text-sm'>
              {t('Slash Progress')}: {activity.cut_count} / {info.max_slashes}
            </p>
            <div className='flex flex-wrap gap-2'>
              <CopyButton
                value={props.affLink || `${window.location.origin}/sign-up`}
                variant='outline'
                size='default'
                aria-label={t('Invite Friends to Slash')}
              >
                {t('Invite Friends to Slash')}
              </CopyButton>
              <Button onClick={() => props.onBuyBargain(activity)}>
                {t('Buy Now at ¥{{price}}', {
                  price: activity.current_price.toFixed(2),
                })}
              </Button>
              {props.onRefresh && (
                <Button variant='outline' onClick={props.onRefresh}>
                  {t('Refresh Progress')}
                </Button>
              )}
            </div>
            {!!info.records?.length && (
              <ul className='max-h-36 overflow-y-auto text-sm'>
                {info.records.map((record) => (
                  <li
                    key={record.id}
                    className='flex justify-between gap-3 py-1'
                  >
                    <span>{record.username}</span>
                    <span>−¥{record.cut_amount.toFixed(2)}</span>
                  </li>
                ))}
              </ul>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}
