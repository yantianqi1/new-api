/*
Copyright (C) 2025 QuantumNous

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

import React, { useContext, useEffect, useState } from 'react';
import {
  Button,
  Typography,
  Input,
  ScrollList,
  ScrollItem,
  Tabs,
  TabPane,
} from '@douyinfe/semi-ui';
import { API, showError, copy, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { API_ENDPOINTS } from '../../constants/common.constant';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import homeHeroBackground from '../../assets/home-hero-background.png';
import {
  IconGithubLogo,
  IconPlay,
  IconFile,
  IconCopy,
  IconKey,
} from '@douyinfe/semi-icons';
import { Link } from 'react-router-dom';
import NoticeModal from '../../components/layout/NoticeModal';

const { Text } = Typography;

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;
  const docsLink = statusState?.status?.docs_link || '';
  const serverAddress = window.location.origin;
  const endpointItems = API_ENDPOINTS.map((e) => ({ value: e }));
  const [endpointIndex, setEndpointIndex] = useState(0);
  const isChinese = i18n.language.startsWith('zh');
  const endpointRows = [
    ['Chat Completions', '/v1/chat/completions'],
    ['Models', '/v1/models'],
    ['Images', '/v1/images/generations'],
  ];
  const quotaUsageUrl = `${serverAddress}/api/usage/token/`;
  const quotaCurlExample = `curl -X GET "${quotaUsageUrl}" \\
  -H "Authorization: Bearer sk-your-api-key"`;
  const quotaFetchExample = `const res = await fetch('${quotaUsageUrl}', {
  headers: {
    Authorization: 'Bearer sk-your-api-key',
  },
});
const { data } = await res.json();
console.log(data.total_available);`;
  const quotaResponseExample = `{
  "code": true,
  "data": {
    "object": "token_usage",
    "name": "default",
    "total_granted": 100000,
    "total_used": 25000,
    "total_available": 75000,
    "unlimited_quota": false,
    "expires_at": 0
  }
}`;
  const quotaAiPrompt = `请帮我在当前项目中集成一个 API Key 额度查询插件/组件。

接口信息：
- 请求方法：GET
- 接口地址：${quotaUsageUrl}
- 鉴权方式：在请求头中传入用户的 API Key
- 请求头格式：Authorization: Bearer sk-your-api-key

返回示例：
${quotaResponseExample}

字段说明：
- data.total_available：当前 API Key 剩余额度
- data.total_used：当前 API Key 已用额度
- data.total_granted：当前 API Key 总额度
- data.unlimited_quota：是否为无限额度
- data.expires_at：过期时间，0 表示永不过期
- data.name：令牌名称

实现要求：
1. 提供一个输入框让用户输入 API Key，不要把 API Key 写死在代码里。
2. 点击查询时调用上述接口，并把 API Key 放到 Authorization 请求头。
3. 成功后展示剩余额度、已用额度、总额度、是否无限额度和有效期。
4. 请求失败时展示清晰错误提示，例如密钥无效、网络错误或接口返回失败。
5. 不要在日志、URL 参数、本地存储或页面明文区域暴露用户输入的完整 API Key。
6. UI 要包含加载状态、防重复点击，以及一键复制查询结果的能力。
7. 如果是在浏览器前端直接调用，请确认目标服务允许跨域；如果不允许，请通过你项目自己的后端代理转发。`;

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);

      // 如果内容是 URL，则发送主题模式
      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  const handleCopyBaseURL = async () => {
    const ok = await copy(serverAddress);
    if (ok) {
      showSuccess(t('已复制到剪切板'));
    }
  };

  const handleCopyText = async (text) => {
    const ok = await copy(text);
    if (ok) {
      showSuccess(t('已复制到剪切板'));
    }
  };

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };

    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      setEndpointIndex((prev) => (prev + 1) % endpointItems.length);
    }, 3000);
    return () => clearInterval(timer);
  }, [endpointItems.length]);

  return (
    <div className='classic-page-fill classic-home-page w-full overflow-x-hidden'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='classic-home-default w-full overflow-x-hidden'>
          <div
            className='classic-home-hero w-full border-b border-semi-color-border relative overflow-hidden'
            style={{ '--classic-home-hero-bg': `url(${homeHeroBackground})` }}
          >
            <div className='classic-home-shell classic-home-hero-grid'>
              <section className='classic-home-hero-copy'>
                <div className='classic-home-eyebrow'>
                  {t('大模型 API 中转站')}
                </div>
                <h1
                  className={`classic-home-title classic-home-title-art ${isChinese ? 'classic-home-title-zh' : ''}`}
                >
                  {t('爱玩Ai')}
                </h1>
                <p className='classic-home-subtitle'>
                  {t(
                    '统一接入多家模型供应商，兼容主流 API 格式，把密钥、额度和用量管理集中到一个入口。',
                  )}
                </p>
                <div className='classic-home-actions'>
                  <Link to='/console'>
                    <Button
                      theme='solid'
                      type='primary'
                      size={isMobile ? 'default' : 'large'}
                      className='classic-home-primary-button'
                      icon={<IconPlay />}
                    >
                      {t('开始使用')}
                    </Button>
                  </Link>
                  {isDemoSiteMode && statusState?.status?.version ? (
                    <Button
                      size={isMobile ? 'default' : 'large'}
                      className='classic-home-secondary-button'
                      icon={<IconGithubLogo />}
                      onClick={() =>
                        window.open(
                          'https://github.com/QuantumNous/new-api',
                          '_blank',
                        )
                      }
                    >
                      {statusState.status.version}
                    </Button>
                  ) : (
                    docsLink && (
                      <Button
                        size={isMobile ? 'default' : 'large'}
                        className='classic-home-secondary-button'
                        icon={<IconFile />}
                        onClick={() => window.open(docsLink, '_blank')}
                      >
                        {t('查看文档')}
                      </Button>
                    )
                  )}
                </div>
                <div className='classic-home-base-block'>
                  <Text className='classic-home-base-label'>
                    {t('当前 Base URL')}
                  </Text>
                  <Input
                    readOnly
                    value={serverAddress}
                    className='classic-home-base-input'
                    size={isMobile ? 'default' : 'large'}
                    suffix={
                      <div className='flex items-center gap-2'>
                        <ScrollList
                          bodyHeight={32}
                          style={{ border: 'unset', boxShadow: 'unset' }}
                        >
                          <ScrollItem
                            mode='wheel'
                            cycled={true}
                            list={endpointItems}
                            selectedIndex={endpointIndex}
                            onSelect={({ index }) => setEndpointIndex(index)}
                          />
                        </ScrollList>
                        <Button
                          type='primary'
                          onClick={handleCopyBaseURL}
                          icon={<IconCopy />}
                          className='classic-home-copy-button'
                          aria-label={t('复制 Base URL')}
                        />
                      </div>
                    }
                  />
                </div>
              </section>

              <aside className='classic-home-preview-card'>
                <div className='classic-home-preview-header'>
                  <div>
                    <Text className='classic-home-preview-kicker'>
                      {t('接入预览')}
                    </Text>
                    <h2>{t('兼容 OpenAI 格式')}</h2>
                  </div>
                  <span>{t('在线')}</span>
                </div>
                <div className='classic-home-code-line'>
                  <span>POST</span>
                  <code>{`${serverAddress}/v1/chat/completions`}</code>
                </div>
                <div className='classic-home-token-line'>
                  <span>Authorization</span>
                  <code>Bearer sk-••••••••••••</code>
                </div>
                <div className='classic-home-endpoints'>
                  {endpointRows.map(([label, path]) => (
                    <div className='classic-home-endpoint-row' key={path}>
                      <span>{label}</span>
                      <code>{path}</code>
                    </div>
                  ))}
                </div>
                <div className='classic-home-preview-tags'>
                  <span>{t('稳定转发')}</span>
                  <span>{t('统一计费')}</span>
                  <span>{t('密钥隔离')}</span>
                </div>
              </aside>
            </div>

            <div className='classic-home-shell classic-home-quota-card'>
              <div className='classic-home-quota-heading'>
                <div className='classic-home-quota-icon'>
                  <IconKey />
                </div>
                <div>
                  <Text className='classic-home-preview-kicker'>
                    {t('开发者接入')}
                  </Text>
                  <h2>{t('额度查询接口')}</h2>
                </div>
                <Button
                  theme='solid'
                  type='primary'
                  icon={<IconCopy />}
                  onClick={() => handleCopyText(quotaAiPrompt)}
                  className='classic-home-quota-button'
                  aria-label={t('复制 AI 集成提示词')}
                >
                  {t('一键复制')}
                </Button>
              </div>
              <Tabs
                className='classic-home-quota-tabs'
                type='button'
                size='small'
              >
                <TabPane tab={t('接口')} itemKey='endpoint'>
                  <div className='classic-home-quota-endpoint'>
                    <span>GET</span>
                    <code>{quotaUsageUrl}</code>
                  </div>
                  <div className='classic-home-quota-result'>
                    <div>
                      <span>{t('认证方式')}</span>
                      <strong>Bearer Token</strong>
                    </div>
                    <div>
                      <span>{t('请求头')}</span>
                      <strong>Authorization</strong>
                    </div>
                    <div>
                      <span>{t('余额字段')}</span>
                      <strong>data.total_available</strong>
                    </div>
                  </div>
                </TabPane>
                <TabPane tab={t('示例')} itemKey='examples'>
                  <div className='classic-home-quota-examples'>
                    <div className='classic-home-quota-example'>
                      <div className='classic-home-quota-example-header'>
                        <span>cURL</span>
                      </div>
                      <pre>{quotaCurlExample}</pre>
                    </div>
                    <div className='classic-home-quota-example'>
                      <div className='classic-home-quota-example-header'>
                        <span>JavaScript</span>
                      </div>
                      <pre>{quotaFetchExample}</pre>
                    </div>
                  </div>
                </TabPane>
                <TabPane tab={t('返回')} itemKey='response'>
                  <div className='classic-home-quota-example classic-home-quota-response'>
                    <div className='classic-home-quota-example-header'>
                      <span>{t('返回示例')}</span>
                    </div>
                    <pre>{quotaResponseExample}</pre>
                  </div>
                </TabPane>
              </Tabs>
            </div>
          </div>
        </div>
      ) : (
        <div className='classic-page-fill overflow-x-hidden w-full'>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              className='w-full h-screen border-none'
            />
          ) : (
            <div
              className='mt-[60px]'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default Home;
