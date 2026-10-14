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
import Claude from '@lobehub/icons/es/Claude/components/Color.js'
import DeepSeek from '@lobehub/icons/es/DeepSeek/components/Color.js'
import Gemini from '@lobehub/icons/es/Gemini/components/Color.js'
import Grok from '@lobehub/icons/es/Grok/components/Mono.js'
import OpenAI from '@lobehub/icons/es/OpenAI/components/Mono.js'
import Qwen from '@lobehub/icons/es/Qwen/components/Color.js'
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  ArrowUpRight,
  Boxes,
  ShieldCheck,
  BarChart3,
  JapaneseYen,
  Zap,
  PawPrint,
  CheckCircle2,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { IconGithub } from '@/assets/brand-icons/icon-github'
import { CopyButton } from '@/components/copy-button'

import ActivityShowcase from './activity-showcase'
import DeveloperSection from './developer-section'
import HeroModelSquare from './hero-model-square'

import '../../waner-home.css'

const providers = [
  { name: 'OpenAI', Icon: OpenAI },
  { name: 'Claude', Icon: Claude },
  { name: 'Gemini', Icon: Gemini },
  { name: 'DeepSeek', Icon: DeepSeek },
  { name: 'Qwen', Icon: Qwen },
  { name: 'Grok', Icon: Grok },
]

interface WanerLandingProps {
  serverAddress?: string
  docsLink?: string
  pricingLink?: string
  demoVersion?: string
  quotaAiPrompt?: string
  isAuthenticated?: boolean
}

export default function WanerLanding({
  serverAddress = typeof window !== 'undefined' ? window.location.origin : '',
  docsLink,
  pricingLink = '/pricing',
  demoVersion,
  quotaAiPrompt,
  isAuthenticated = false,
}: WanerLandingProps) {
  const { t } = useTranslation()

  const isPricingExternal =
    Boolean(pricingLink) && /^(https?:)?\/\//i.test(pricingLink)
  const targetPricingUrl = pricingLink || '/pricing'

  const capabilities = [
    {
      title: t('One API, many models'),
      body: t(
        'Reach every configured upstream provider through a single OpenAI-compatible endpoint.'
      ),
      icon: Boxes,
      iconTheme: 'waner-card-orange',
    },
    {
      title: t('Keys you control'),
      body: t(
        'Issue a key per project, set a spending limit, and revoke access whenever you need to.'
      ),
      icon: ShieldCheck,
      iconTheme: 'waner-card-green',
    },
    {
      title: t('Usage you can audit'),
      body: t(
        'Each call is recorded with its model, token counts, and cost, so spending stays reconcilable.'
      ),
      icon: BarChart3,
      iconTheme: 'waner-card-pink',
    },
    {
      title: t('Pricing shown up front'),
      body: t(
        'Model prices and multipliers follow the marketplace configuration rather than this page.'
      ),
      icon: JapaneseYen,
      iconTheme: 'waner-card-purple',
    },
    {
      title: t('Drop-in compatible'),
      body: t(
        'Point an existing OpenAI SDK at the base URL. Some clients need one additional setting.'
      ),
      icon: Zap,
      iconTheme: 'waner-card-blue',
    },
    {
      title: t('Your own gateway'),
      body: t(
        'Open source and self-hosted: the deployment, the data, and the channels remain yours.'
      ),
      icon: PawPrint,
      iconTheme: 'waner-card-yellow',
    },
  ]

  const primaryTarget = isAuthenticated ? '/dashboard' : '/sign-up'

  return (
    <div className='home-v2 waner-home'>
      {/* ── 1. Hero Section ────────────────────────────────────────── */}
      <section className='waner-hero' aria-labelledby='waner-hero-title'>
        <div className='waner-hero-shell'>
          {/* Left Column: Copy & Actions */}
          <div className='waner-hero-copy'>
            <div className='waner-eyebrow-tag'>
              <span className='waner-tag-brand'>waner~API站</span>
              <span className='waner-tag-divider' aria-hidden='true' />
              <span className='waner-tag-slogan'>
                {t('Bridge connecting AI and creativity')}
              </span>
            </div>

            <h1 id='waner-hero-title' className='waner-hero-headline'>
              <span className='waner-headline-line'>{t('Great ideas.')}</span>
              <span className='waner-headline-line'>
                {t('Connected to AI.')}
                <span className='waner-headline-paw' aria-hidden='true'>
                  🐾
                </span>
              </span>
              <span className='waner-headline-script' aria-hidden='true'>
                waner~
              </span>
            </h1>

            <p className='waner-hero-lede'>
              {t(
                'A stable, easy-to-use, and transparently priced API relay station. Connect to multiple AI models, write code, build products, and explore creativity — start here.'
              )}
            </p>

            <div className='waner-hero-actions'>
              <Link className='waner-btn waner-btn-primary' to={primaryTarget}>
                <PawPrint size={18} aria-hidden='true' />
                <span>
                  {isAuthenticated ? t('Enter Console') : t('Start using')}
                </span>
                <ArrowRight size={17} aria-hidden='true' />
              </Link>

              {isPricingExternal ? (
                <a
                  className='waner-btn waner-btn-secondary'
                  href={targetPricingUrl}
                  target='_blank'
                  rel='noopener noreferrer'
                >
                  <span>{t('Explore models')}</span>
                  <ArrowRight size={17} aria-hidden='true' />
                </a>
              ) : (
                <Link className='waner-btn waner-btn-secondary' to='/pricing'>
                  <span>{t('Explore models')}</span>
                  <ArrowRight size={17} aria-hidden='true' />
                </Link>
              )}
            </div>

            <div className='waner-endpoint-card'>
              <span className='waner-endpoint-label'>{t('API endpoint')}</span>
              <div className='waner-endpoint-field'>
                <code translate='no'>{serverAddress}</code>
                <CopyButton
                  value={serverAddress}
                  aria-label={t('Copy Base URL')}
                  tooltip={t('Copy Base URL')}
                  variant='outline'
                  size='icon'
                />
              </div>

              <div className='waner-trust-pills'>
                <span className='waner-trust-pill waner-pill-green'>
                  <CheckCircle2 size={13} aria-hidden='true' />
                  {t('Stable & Fast')}
                </span>
                <span className='waner-trust-pill waner-pill-orange'>
                  <CheckCircle2 size={13} aria-hidden='true' />
                  {t('Transparent Pricing')}
                </span>
                <span className='waner-trust-pill waner-pill-green'>
                  <CheckCircle2 size={13} aria-hidden='true' />
                  {t('Multi-model Support')}
                </span>
                <span className='waner-trust-pill waner-pill-orange'>
                  <CheckCircle2 size={13} aria-hidden='true' />
                  {t('Beginner Friendly')}
                </span>
              </div>
            </div>

            <div className='waner-easel-note' aria-hidden='true'>
              <img
                src='/img/home/easel_card.png'
                alt='Life is better with AI wer~'
                loading='lazy'
                onError={(e) => {
                  e.currentTarget.style.display = 'none'
                }}
              />
            </div>
          </div>

          {/* Right Column: Hero Visual Artwork & Live Model Square */}
          <div className='waner-hero-visual'>
            <div className='waner-visual-container'>
              <img
                src='/img/home/hero_visual.png'
                alt='waner~ puppy mascot with laptop and model square'
                className='waner-visual-img'
                loading='eager'
                onError={(e) => {
                  e.currentTarget.style.display = 'none'
                }}
              />

              <HeroModelSquare />
            </div>
          </div>
        </div>
      </section>

      {/* ── 2. Activities Carousel Section (精选活动：抽奖/邀请/档位/免费) ─── */}
      <ActivityShowcase />

      {/* ── 3. Features Section ("路由清晰，价格实时，记录可查") ────── */}
      <section
        className='waner-features-section'
        id='models'
        aria-labelledby='waner-features-title'
      >
        <div className='waner-shell'>
          <div className='waner-features-head'>
            <div className='waner-head-accent-left' aria-hidden='true'>
              <span className='waner-paw-icon'>🐾</span>
            </div>

            <div className='waner-head-text'>
              <h2 id='waner-features-title' className='waner-features-title'>
                {t('Clear routing, live pricing, records you can check')}
              </h2>
              <p className='waner-features-subtitle'>
                {t(
                  'Make every call simpler, more transparent, and more reassuring.'
                )}
              </p>
            </div>

            <div className='waner-head-accent-right' aria-hidden='true'>
              <img
                src='/img/home/running_puppy.png'
                alt='wer~ 向美好的 AI 出发！'
                className='waner-running-dog'
                loading='lazy'
                onError={(e) => {
                  e.currentTarget.style.display = 'none'
                }}
              />
            </div>
          </div>

          <div className='waner-cards-grid'>
            {capabilities.map(({ title, body, icon: Icon, iconTheme }) => (
              <div key={title} className='waner-feature-card'>
                <div className={`waner-card-icon-box ${iconTheme}`}>
                  <Icon size={24} aria-hidden='true' />
                </div>
                <h3 className='waner-card-title'>{title}</h3>
                <p className='waner-card-body'>{body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── 4. Model Ecosystem Floating Bar ("模型生态") ────────────── */}
      <section
        className='waner-ecosystem-section'
        aria-label={t('Model ecosystem')}
      >
        <div className='waner-shell waner-ecosystem-wrapper'>
          <div className='waner-ecosystem-bar'>
            <span className='waner-ecosystem-label'>
              {t('Model ecosystem')}
            </span>

            <ul className='waner-ecosystem-list'>
              {providers.map(({ name, Icon }) => (
                <li key={name} className='waner-ecosystem-item'>
                  <Icon size={20} aria-hidden='true' />
                  <span translate='no'>{name}</span>
                </li>
              ))}
            </ul>

            <Link className='waner-ecosystem-link' to='/pricing'>
              <span>{t('View models and pricing')}</span>
              <ArrowRight size={15} aria-hidden='true' />
            </Link>
          </div>

          <div className='waner-sleeping-dog-box' aria-hidden='true'>
            <img
              src='/img/home/sleeping_puppy.png'
              alt='Powering Your Ideas wer~'
              className='waner-sleeping-dog'
              loading='lazy'
              onError={(e) => {
                e.currentTarget.style.display = 'none'
              }}
            />
          </div>
        </div>
      </section>

      {/* ── 5. Developer Quickstart Section ────────────────────────── */}
      <DeveloperSection
        serverAddress={serverAddress}
        docsLink={docsLink}
        quotaAiPrompt={quotaAiPrompt}
      />

      {/* ── 6. Closing Section ─────────────────────────────────────── */}
      <section
        className='waner-closing-section'
        aria-labelledby='waner-closing-title'
      >
        <div className='waner-shell waner-closing-inner'>
          <div className='waner-closing-text'>
            <h2 id='waner-closing-title'>
              {t('Your next great thing.')}
              <span className='waner-closing-highlight'>
                {t('It starts here.')}
              </span>
            </h2>
            <p>
              {t(
                'Give your ideas a place to begin. We will handle the connection.'
              )}
            </p>
          </div>
          <div className='waner-closing-actions'>
            <Link className='waner-btn waner-btn-primary' to={primaryTarget}>
              <PawPrint size={18} aria-hidden='true' />
              <span>{t('Start building')}</span>
              <ArrowUpRight size={17} aria-hidden='true' />
            </Link>
            <div className='waner-closing-links'>
              {docsLink && (
                <a
                  href={docsLink}
                  target='_blank'
                  rel='noopener noreferrer'
                  className='waner-closing-link'
                >
                  <span>{t('Read the documentation')}</span>
                  <ArrowUpRight size={15} aria-hidden='true' />
                </a>
              )}
              {demoVersion && (
                <a
                  href='https://github.com/QuantumNous/new-api'
                  target='_blank'
                  rel='noopener noreferrer'
                  className='waner-closing-link'
                >
                  <IconGithub className='size-4' aria-hidden='true' />
                  <span>{demoVersion}</span>
                </a>
              )}
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}
