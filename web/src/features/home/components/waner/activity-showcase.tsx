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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

import { HOME_ACTIVITIES } from './activity-data'

const AUTOPLAY_INTERVAL = 8000

export default function ActivityShowcase() {
  const { t } = useTranslation()
  const [activeIndex, setActiveIndex] = useState(0)
  const [isPaused, setIsPaused] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const total = HOME_ACTIVITIES.length
  const current = HOME_ACTIVITIES[activeIndex]

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
      className='waner-showcase-section'
      aria-label={t('Featured Activities')}
    >
      <div className='waner-shell'>
        {/* Reuse the project tabs for keyboard navigation and active panels. */}
        <Tabs
          value={HOME_ACTIVITIES[activeIndex].id}
          onValueChange={(value) => {
            const index = HOME_ACTIVITIES.findIndex(
              (activity) => activity.id === value
            )
            if (index >= 0) setActiveIndex(index)
          }}
        >
          <div className='waner-showcase-header'>
            <h2 className='waner-showcase-heading'>
              {t('Featured Activities')}
            </h2>
            <TabsList
              className='waner-showcase-tabs home-activities-tabs'
              aria-label={t('Featured Activities')}
            >
              {HOME_ACTIVITIES.map((act, index) => (
                <TabsTrigger
                  key={act.id}
                  value={act.id}
                  className={`waner-showcase-tab home-activity-tab ${index === activeIndex ? 'active' : ''} ${act.theme}`}
                >
                  <span
                    aria-hidden='true'
                    className='waner-tab-icon home-activity-tab-icon'
                  >
                    {act.tabIcon}
                  </span>
                  <span className='waner-tab-name'>{t(act.tabName)}</span>
                </TabsTrigger>
              ))}
            </TabsList>
          </div>

          {/* Carousel Card */}
          <TabsContent
            value={current.id}
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
                    <span
                      aria-hidden='true'
                      className='waner-feat-icon home-act-bullet-icon'
                    >
                      {feat.icon}
                    </span>
                    <span className='waner-feat-text'>{t(feat.text)}</span>
                  </li>
                ))}
              </ul>

              <div className='waner-showcase-actions home-activities-controls'>
                <Button
                  nativeButton={false}
                  role='link'
                  render={
                    <a
                      href={current.href}
                      target='_blank'
                      rel='noopener noreferrer'
                    />
                  }
                  size='lg'
                  variant='ghost'
                  className='waner-showcase-cta home-act-btn'
                >
                  <span>{t(current.ctaText)}</span>
                  <ArrowUpRight size={17} aria-hidden='true' />
                </Button>

                <div className='waner-carousel-controls home-activities-arrows'>
                  <Button
                    variant='outline'
                    size='icon'
                    type='button'
                    onClick={goPrev}
                    className='waner-carousel-btn home-activities-arrow'
                    aria-label={t('Previous activity')}
                  >
                    <ChevronLeft size={18} />
                  </Button>
                  <span className='waner-carousel-count'>
                    {activeIndex + 1} / {total}
                  </span>
                  <Button
                    variant='outline'
                    size='icon'
                    type='button'
                    onClick={goNext}
                    className='waner-carousel-btn home-activities-arrow'
                    aria-label={t('Next activity')}
                  >
                    <ChevronRight size={18} />
                  </Button>
                </div>
              </div>
            </div>

            <div className='waner-showcase-banner home-act-banner'>
              {current.bannerType === 'composite-models' && (
                <div className='waner-models-visual home-act-models-layout'>
                  <div className='home-act-models-bg' aria-hidden='true' />
                  <img
                    src='/img/activities/models/char_left.png'
                    alt=''
                    className='home-act-char-left'
                    aria-hidden='true'
                  />
                  <img
                    src={current.bannerSrc}
                    alt={t(current.title)}
                    className='home-act-husky-center'
                  />
                  <img
                    src='/img/activities/models/char_right.png'
                    alt=''
                    className='home-act-char-right'
                    aria-hidden='true'
                  />
                </div>
              )}
              {current.bannerType === 'picture' && (
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
              )}
              {current.bannerType === 'image' && (
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
          </TabsContent>
        </Tabs>
      </div>
    </section>
  )
}
