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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useStatus } from '@/hooks/use-status'

import './style.css'

export function ActivityBar() {
  const { t } = useTranslation()
  const { status } = useStatus()
  if (status?.activity_bar_enabled !== true) return null
  let rawItems = status.activity_bar_items
  if (typeof rawItems === 'string') {
    try {
      rawItems = JSON.parse(rawItems)
    } catch {
      return null
    }
  }
  if (!Array.isArray(rawItems)) return null
  const items: {
    title: string
    link?: string
    tag?: string
    external: boolean
  }[] = []
  for (const raw of rawItems) {
    if (
      !raw ||
      typeof raw !== 'object' ||
      typeof raw.title !== 'string' ||
      !raw.title.trim()
    ) {
      continue
    }
    const link = typeof raw.link === 'string' ? raw.link.trim() : undefined
    const external = Boolean(link && /^https?:\/\//i.test(link))
    const internal = Boolean(
      link &&
      link.startsWith('/') &&
      !link.startsWith('//') &&
      !link.includes('\\')
    )
    if (link && !external && !internal) continue
    items.push({
      title: raw.title,
      link,
      tag: typeof raw.tag === 'string' ? raw.tag : undefined,
      external,
    })
  }
  if (!items.length) return null
  return (
    <nav
      aria-label={t('Activity Bar')}
      className='activity-bar border-border bg-primary/5 w-full overflow-hidden border-b text-sm'
    >
      <div className='activity-bar-track flex w-max items-center py-2'>
        <div className='flex items-center gap-8 px-4'>
          {items.map((item) => {
            const content = (
              <>
                {t(item.title)}{' '}
                {item.tag && (
                  <span className='bg-primary/10 text-primary rounded px-1.5 py-0.5 text-xs'>
                    {t(item.tag)}
                  </span>
                )}
              </>
            )
            const className =
              'inline-flex items-center gap-2 whitespace-nowrap focus-visible:outline-2 focus-visible:outline-offset-2'
            if (item.external) {
              return (
                <a
                  key={`${item.title}:${item.link}:${item.tag}`}
                  className={className}
                  href={item.link}
                  target='_blank'
                  rel='noopener noreferrer'
                >
                  {content}
                </a>
              )
            }
            if (item.link) {
              return (
                <Link
                  key={`${item.title}:${item.link}:${item.tag}`}
                  className={className}
                  to={item.link}
                >
                  {content}
                </Link>
              )
            }
            return (
              <span
                key={`${item.title}:${item.link}:${item.tag}`}
                className={className}
              >
                {content}
              </span>
            )
          })}
        </div>
        <div
          aria-hidden='true'
          className='activity-bar-repeat flex items-center gap-8 px-4'
        >
          {items.map((item) => (
            <span
              key={`${item.title}:${item.link}:${item.tag}`}
              className='inline-flex items-center gap-2 whitespace-nowrap'
            >
              {t(item.title)}
              {item.tag && (
                <span className='bg-primary/10 text-primary rounded px-1.5 py-0.5 text-xs'>
                  {t(item.tag)}
                </span>
              )}
            </span>
          ))}
        </div>
      </div>
    </nav>
  )
}
