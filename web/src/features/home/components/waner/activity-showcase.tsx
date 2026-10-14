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
import { ArrowUpRight, ChevronLeft, ChevronRight } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

const HUB_URL = 'https://gift.inkwan.app'

interface ActivityFeature {
  icon: string
  text: string
}

interface ActivityItem {
  id: string
  href: string
  theme: string
  tabName: string
  tabIcon: string
  badge: string
  kicker: string
  title: string
  subtitle: string
  features: ActivityFeature[]
  ctaText: string
  bannerType: 'image' | 'picture'
  bannerSrc: string
  bannerWebp?: string
  bannerFallback: string
}

const ACTIVITIES: ActivityItem[] = [
  {
    id: 'lottery',
    href: `${HUB_URL}/lottery`,
    theme: 'theme-lottery',
    tabName: 'Lucky draw',
    tabIcon: '🎁',
    badge: 'Instant draw on recharge',
    kicker: 'LUCKY DRAW · ROULETTE',
    title: 'Recharge lucky draw',
    subtitle:
      'Hanako-kun supernatural theme! Single recharge of ¥10+ earns 1 free lucky wheel draw with full refunds and real quotas.',
    features: [
      { icon: '🎯', text: 'Single qualifying recharge earns 1 draw chance' },
      {
        icon: '👑',
        text: 'Chance to win full-refund equivalent codes and cash quotas',
      },
      {
        icon: '⚡',
        text: 'Direct integration with API gateway to auto-generate redemption keys',
      },
    ],
    ctaText: 'Enter draw now',
    bannerType: 'image',
    bannerSrc: '/img/activities/lottery_hero.png',
    bannerFallback: 'https://gift.inkwan.app/img/lottery_hero.png',
  },
  {
    id: 'invite',
    href: `${HUB_URL}/invite`,
    theme: 'theme-invite',
    tabName: 'Invite',
    tabIcon: '👤',
    badge: '10% Lifetime rebate',
    kicker: 'AFFILIATE · PROGRAM',
    title: 'Developer Affiliate Rebate',
    subtitle:
      'Share your referral link with friends and get a 10% quota rebate in seconds upon their recharge. Permanent lock-in, universal across models!',
    features: [
      {
        icon: '💎',
        text: '10% recharge amount credited instantly as real quota',
      },
      {
        icon: '🔗',
        text: 'One-click account binding to generate your exclusive referral link',
      },
      {
        icon: '🚀',
        text: 'Site-wide model compute deduction with no threshold',
      },
    ],
    ctaText: 'Get referral link',
    bannerType: 'picture',
    bannerWebp: '/img/activities/invite/hero_invite.webp',
    bannerSrc: '/img/activities/invite/hero_invite.png',
    bannerFallback: 'https://gift.inkwan.app/img/invite/hero_invite.png',
  },
  {
    id: 'tiers',
    href: `${HUB_URL}/tiers`,
    theme: 'theme-tiers',
    tabName: 'Special Tiers',
    tabIcon: '⚡',
    badge: 'Recharge 5 get 5 bonus',
    kicker: 'SPECIAL TIERS · BONUS',
    title: 'Recharge Special Bonus Tiers',
    subtitle:
      'Limited-time compute carnival! Recharge 5 get 5 with 100% bonus, recharge 100 get 15, recharge 200 get 40. Universal across 40+ models!',
    features: [
      {
        icon: '🎉',
        text: 'First recharge 100% bonus: spend ¥5 to get ¥10 equivalent compute',
      },
      {
        icon: '🎁',
        text: 'Tiered extra rewards: ¥100 gets ¥115, ¥200 gets ¥240',
      },
      {
        icon: '🔥',
        text: 'No threshold usage, models across the station immediately available',
      },
    ],
    ctaText: 'View recharge tiers',
    bannerType: 'picture',
    bannerWebp: '/img/activities/tiers_hero.webp',
    bannerSrc: '/img/activities/tiers_hero.jpg',
    bannerFallback: 'https://gift.inkwan.app/img/tiers/hero_tiers.png',
  },
  {
    id: 'free',
    href: `${HUB_URL}/models`,
    theme: 'theme-models',
    tabName: 'Free Models',
    tabIcon: '✨',
    badge: '0-cost daily quota',
    kicker: 'FREE MODELS · COMMUNITY',
    title: 'Free Daily Compute Models',
    subtitle:
      'Community exclusive benefits! Free access to selected high-performance models every day with zero threshold. Start learning and developing now!',
    features: [
      {
        icon: '🆓',
        text: 'Every registered account gets daily free model quota',
      },
      {
        icon: '⚡',
        text: 'Selected high-speed Gemini and deep-reasoning models free forever',
      },
      {
        icon: '🛠️',
        text: 'Seamless drop-in switch, identical API protocol to paid models',
      },
    ],
    ctaText: 'View free models',
    bannerType: 'image',
    bannerSrc: '/img/activities/models/title_banner.png',
    bannerFallback: 'https://gift.inkwan.app/img/models_hero.png',
  },
]

const AUTOPLAY_INTERVAL = 8000

export default function ActivityShowcase() {
  const { t } = useTranslation()
  const [activeIndex, setActiveIndex] = useState(0)
  const [isPaused, setIsPaused] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const total = ACTIVITIES.length
  const current = ACTIVITIES[activeIndex]

  const goNext = useCallback(() => {
    setActiveIndex((prev) => (prev + 1) % total)
  }, [total])

  const goPrev = useCallback(() => {
    setActiveIndex((prev) => (prev - 1 + total) % total)
  }, [total])

  useEffect(() => {
    if (isPaused) {
      if (timerRef.current) clearInterval(timerRef.current)
      return
    }
    timerRef.current = setInterval(goNext, AUTOPLAY_INTERVAL)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [isPaused, goNext])

  return (
    <section
      className='waner-showcase-section home-capabilities'
      aria-label={t('Featured Activities')}
    >
      <div className='waner-shell'>
        {/* Navigation Tabs */}
        <Tabs
          value={ACTIVITIES[activeIndex].id}
          onValueChange={(value) => {
            const index = ACTIVITIES.findIndex(
              (activity) => activity.id === value
            )
            if (index >= 0) setActiveIndex(index)
          }}
        >
          <TabsList
            className='waner-showcase-tabs home-activities-tabs'
            aria-label={t('Featured Activities')}
          >
            {ACTIVITIES.map((act, index) => (
              <TabsTrigger
                key={act.id}
                value={act.id}
                className={`waner-showcase-tab home-activity-tab ${index === activeIndex ? 'active' : ''} ${act.theme}`}
              >
                <span className='waner-tab-icon home-activity-tab-icon'>
                  {act.tabIcon}
                </span>
                <span className='waner-tab-name'>{t(act.tabName)}</span>
                {index === activeIndex && (
                  <span className='waner-tab-pill'>{t(act.badge)}</span>
                )}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>

        {/* Carousel Card */}
        <div
          className={`waner-showcase-card home-activities-window home-act-card ${current.theme}`}
          onMouseEnter={() => setIsPaused(true)}
          onMouseLeave={() => setIsPaused(false)}
          onFocus={() => setIsPaused(true)}
          onBlur={(event) => {
            if (!event.currentTarget.contains(event.relatedTarget)) {
              setIsPaused(false)
            }
          }}
        >
          <div className='waner-showcase-content home-act-body'>
            <div className='waner-showcase-badge-row home-act-header'>
              <span className='waner-showcase-kicker home-act-kicker'>
                {t(current.kicker)}
              </span>
              <span className='waner-showcase-tag home-act-badge'>
                {t(current.badge)}
              </span>
            </div>

            <h3 className='waner-showcase-title home-act-title'>
              {t(current.title)}
            </h3>
            <p className='waner-showcase-subtitle home-act-desc'>
              {t(current.subtitle)}
            </p>

            <ul className='waner-showcase-features home-act-features'>
              {current.features.map((feat) => (
                <li key={feat.text} className='waner-feature-item'>
                  <span className='waner-feat-icon home-act-bullet-icon'>
                    {feat.icon}
                  </span>
                  <span className='waner-feat-text'>{t(feat.text)}</span>
                </li>
              ))}
            </ul>

            <div className='waner-showcase-actions home-activities-controls'>
              <a
                href={current.href}
                target='_blank'
                rel='noopener noreferrer'
                className='waner-showcase-cta home-act-btn'
              >
                <span>{t(current.ctaText)}</span>
                <ArrowUpRight size={17} aria-hidden='true' />
              </a>

              <div className='waner-carousel-controls home-activities-arrows'>
                <Button
                  variant='outline'
                  size='icon'
                  type='button'
                  onClick={goPrev}
                  className='waner-carousel-btn home-activities-arrow'
                  aria-label={t('Previous')}
                >
                  <ChevronLeft size={18} />
                </Button>
                <span className='waner-carousel-count home-activities-count'>
                  {activeIndex + 1} / {total}
                </span>
                <Button
                  variant='outline'
                  size='icon'
                  type='button'
                  onClick={goNext}
                  className='waner-carousel-btn home-activities-arrow'
                  aria-label={t('Next')}
                >
                  <ChevronRight size={18} />
                </Button>
              </div>
            </div>
          </div>

          <div className='waner-showcase-banner home-act-banner'>
            {current.bannerType === 'picture' && current.bannerWebp ? (
              <picture>
                <source srcSet={current.bannerWebp} type='image/webp' />
                <img
                  src={current.bannerSrc}
                  alt={t(current.title)}
                  className='waner-banner-img home-act-banner-img'
                  onError={(e) => {
                    e.currentTarget.src = current.bannerFallback
                  }}
                />
              </picture>
            ) : (
              <img
                src={current.bannerSrc}
                alt={t(current.title)}
                className='waner-banner-img home-act-banner-img'
                onError={(e) => {
                  e.currentTarget.src = current.bannerFallback
                }}
              />
            )}
          </div>
        </div>
      </div>
    </section>
  )
}
