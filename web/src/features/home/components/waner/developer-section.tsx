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
import { ArrowUpRight, KeyRound } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import CodeSamples, { type CodeExample } from './code-samples'
import CopyButton from './copy-button'

interface DeveloperSectionProps {
  serverAddress: string
  docsLink?: string
  quotaAiPrompt?: string
}

export default function DeveloperSection({
  serverAddress,
  docsLink,
  quotaAiPrompt,
}: DeveloperSectionProps) {
  const { t } = useTranslation()
  const usageUrl = `${serverAddress}/api/usage/token/`
  const defaultAiPrompt = t(
    'Help me integrate an API key quota lookup component into this project.\nUse GET {{usageUrl}} with the Authorization: Bearer <API_KEY> header.\nThe response includes total_granted (total quota), total_used (used quota), and total_available (remaining quota).\nUse my project’s technology stack to build a clear lookup and display module with error handling.',
    { usageUrl }
  )

  const promptText = quotaAiPrompt || defaultAiPrompt

  const examples: CodeExample[] = [
    {
      label: 'cURL',
      language: 'bash',
      code: `curl '${serverAddress}/v1/chat/completions' \\
  -H "Authorization: Bearer $WANER_API_KEY" \\
  -H 'Content-Type: application/json' \\
  -d '{
    "model": "YOUR_MODEL_NAME",
    "messages": [
      { "role": "user", "content": "Hello, AI." }
    ]
  }'`,
      note: t(
        'Set WANER_API_KEY and replace YOUR_MODEL_NAME with a model from the marketplace.'
      ),
    },
    {
      label: 'Python',
      language: 'python',
      code: `import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["WANER_API_KEY"],
    base_url="${serverAddress}/v1",
)

response = client.chat.completions.create(
    model="YOUR_MODEL_NAME",
    messages=[
        {"role": "user", "content": "Hello, AI."},
    ],
)
print(response.choices[0].message.content)`,
      note: t(
        'Requires openai>=1.0. Point base_url at the relay root with /v1 appended.'
      ),
    },
    {
      label: 'Node.js',
      language: 'javascript',
      code: `import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: process.env.WANER_API_KEY,
  baseURL: '${serverAddress}/v1',
});

const response = await client.chat.completions.create({
  model: 'YOUR_MODEL_NAME',
  messages: [{ role: 'user', content: 'Hello, AI.' }],
});
console.log(response.choices[0].message.content);`,
      note: t(
        'Use the official openai npm package. Keep your key in process.env.'
      ),
    },
  ]

  return (
    <section
      className='waner-developer-section home-developer-section'
      aria-labelledby='waner-dev-title'
    >
      <div className='waner-shell'>
        <div className='waner-developer-head home-developer-copy'>
          <div className='waner-developer-eyebrow'>
            <span className='waner-eyebrow-dot' aria-hidden='true' />
            <span>{t('Built for developers')}</span>
          </div>

          <h2 id='waner-dev-title' className='waner-developer-title'>
            {t('Drop-in OpenAI SDK compatibility')}
          </h2>
          <p className='waner-developer-lede home-section-description'>
            {t(
              'Switch from OpenAI or another gateway in seconds. Change the base URL and API key; the payload, streaming behavior, and tool calls work the same.'
            )}
          </p>
        </div>

        <div className='waner-developer-grid home-developer-grid'>
          {/* Main API call sample */}
          <div className='waner-dev-card waner-dev-card-main home-developer-example home-quota-info'>
            <div className='waner-dev-card-head home-example-heading'>
              <div className='waner-dev-card-title-group'>
                <span className='waner-dev-card-badge'>
                  {t('Chat completions')}
                </span>
                <h3 className='waner-dev-card-title'>
                  {t('OpenAI format calls')}
                </h3>
              </div>
              <div className='waner-dev-card-meta'>
                <code>POST /v1/chat/completions</code>
              </div>
            </div>

            <CodeSamples
              examples={examples}
              label={t('Code examples for chat completions')}
            />
          </div>

          {/* Side: Quota & Token Management */}
          <div className='waner-dev-card waner-dev-card-side home-quota-info'>
            <div className='waner-dev-card-head home-example-heading'>
              <div className='waner-dev-card-title-group'>
                <span className='waner-dev-card-badge waner-badge-teal'>
                  {t('Quota endpoint')}
                </span>
                <h3 className='waner-dev-card-title'>
                  {t('Check key balance')}
                </h3>
              </div>
            </div>

            <p className='waner-dev-side-lede home-section-description'>
              {t(
                'Let front-end clients, bots, and agents inspect their remaining allowance directly.'
              )}
            </p>

            <div className='waner-endpoint-pill home-inline-endpoint'>
              <div className='waner-endpoint-pill-top'>
                <span className='waner-pill-method'>GET</span>
                <code translate='no' className='waner-pill-url'>
                  {usageUrl}
                </code>
              </div>
              <div className='waner-endpoint-pill-actions'>
                <CopyButton
                  text={usageUrl}
                  label={t('Copy token quota check URL')}
                  compact
                />
              </div>
            </div>

            <div className='waner-ai-prompt-box'>
              <div className='waner-ai-prompt-head'>
                <div className='waner-ai-prompt-label'>
                  <span className='waner-sparkle' aria-hidden='true'>
                    ✨
                  </span>
                  <span>{t('AI Prompt for integrating balance check')}</span>
                </div>
                <CopyButton
                  text={promptText}
                  label={t('Copy AI integration prompt')}
                  compact
                />
              </div>
              <p className='waner-ai-prompt-text'>{promptText}</p>
            </div>

            <div className='waner-dev-side-foot'>
              <Link to='/keys' className='waner-dev-side-btn'>
                <KeyRound size={15} aria-hidden='true' />
                <span>{t('Manage API Keys')}</span>
              </Link>
              {docsLink && (
                <a
                  href={docsLink}
                  target='_blank'
                  rel='noopener noreferrer'
                  className='waner-dev-side-link'
                >
                  <span>{t('Documentation')}</span>
                  <ArrowUpRight size={14} aria-hidden='true' />
                </a>
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
