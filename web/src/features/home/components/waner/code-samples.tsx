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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CodeBlock } from '@/components/ai-elements/code-block'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'

import CopyButton from './copy-button'

export interface CodeExample {
  label: string
  language: string
  code: string
  note?: string
}
interface CodeSamplesProps {
  examples: CodeExample[]
  label: string
}
export default function CodeSamples(props: CodeSamplesProps) {
  const { t } = useTranslation()
  const [activeLabel, setActiveLabel] = useState(props.examples[0]?.label ?? '')
  const active =
    props.examples.find((example) => example.label === activeLabel) ??
    props.examples[0]
  if (!active) return null
  return (
    <Tabs
      value={active.label}
      onValueChange={(value) => setActiveLabel(String(value))}
      className='home-code-samples'
    >
      <div className='home-code-toolbar'>
        <TabsList aria-label={props.label}>
          {props.examples.map((example) => (
            <TabsTrigger key={example.label} value={example.label}>
              {example.label}
            </TabsTrigger>
          ))}
        </TabsList>
        <CopyButton
          text={active.code}
          label={t('Copy {{language}} sample', { language: active.label })}
          compact
        />
      </div>
      {props.examples.map((example) => (
        <TabsContent key={example.label} value={example.label}>
          <CodeBlock
            code={example.code}
            language={example.language}
            showToolbar={false}
          />
          {example.note && <p className='home-code-note'>{example.note}</p>}
        </TabsContent>
      ))}
    </Tabs>
  )
}
