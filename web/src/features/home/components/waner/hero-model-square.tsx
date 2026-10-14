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
import OpenAI from '@lobehub/icons/es/OpenAI/components/Mono.js'
import Qwen from '@lobehub/icons/es/Qwen/components/Color.js'
import { Link } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

const featuredModels = [
  { name: 'GPT-6', provider: 'OpenAI', Icon: OpenAI },
  { name: 'Claude Opus 5.5', provider: 'Anthropic', Icon: Claude },
  { name: 'Gemini 4', provider: 'Google', Icon: Gemini },
  { name: 'DeepSeek V4.1 Flash', provider: 'DeepSeek', Icon: DeepSeek },
  { name: 'Qwen 3.8-Max', provider: 'Qwen', Icon: Qwen },
]

export default function HeroModelSquare() {
  const { t } = useTranslation()

  return (
    <nav className='waner-hero-model-square' aria-label={t('Model Square')}>
      <Link to='/pricing' className='waner-hero-model-heading'>
        <span>{t('Model Square')}</span>
        <ArrowRight aria-hidden='true' />
      </Link>
      <ul className='waner-hero-model-list'>
        {featuredModels.map((model) => (
          <li key={model.name}>
            <Link
              to='/pricing'
              className='waner-hero-model-link'
              aria-label={model.name}
            >
              <model.Icon
                className='waner-hero-model-icon'
                aria-hidden='true'
              />
              <span className='waner-hero-model-copy' translate='no'>
                <strong>{model.name}</strong>
                <span>{model.provider}</span>
              </span>
            </Link>
          </li>
        ))}
      </ul>
      <Link to='/pricing' className='waner-hero-model-footer'>
        <span>{t('View all models')}</span>
        <ArrowRight aria-hidden='true' />
      </Link>
    </nav>
  )
}
