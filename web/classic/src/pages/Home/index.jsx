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
} from '@douyinfe/semi-ui';
import { API, showError, copy, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { API_ENDPOINTS } from '../../constants/common.constant';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import {
  IconGithubLogo,
  IconPlay,
  IconFile,
  IconCopy,
} from '@douyinfe/semi-icons';
import { Link } from 'react-router-dom';
import NoticeModal from '../../components/layout/NoticeModal';
import {
  Moonshot,
  OpenAI,
  XAI,
  Zhipu,
  Volcengine,
  Cohere,
  Claude,
  Gemini,
  Suno,
  Minimax,
  Wenxin,
  Spark,
  Qingyan,
  DeepSeek,
  Qwen,
  Midjourney as MjProxyIcon,
  Grok,
  AzureAI,
  Hunyuan,
  Xinference,
} from '@lobehub/icons';

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
  const providerIcons = [
    <Moonshot size={40} />,
    <OpenAI size={40} />,
    <XAI size={40} />,
    <Zhipu.Color size={40} />,
    <Volcengine.Color size={40} />,
    <Cohere.Color size={40} />,
    <Claude.Color size={40} />,
    <Gemini.Color size={40} />,
    <Suno size={40} />,
    <Minimax.Color size={40} />,
    <Wenxin.Color size={40} />,
    <Spark.Color size={40} />,
    <Qingyan.Color size={40} />,
    <DeepSeek.Color size={40} />,
    <Qwen.Color size={40} />,
    <MjProxyIcon size={40} />,
    <Grok size={40} />,
    <AzureAI.Color size={40} />,
    <Hunyuan.Color size={40} />,
    <Xinference.Color size={40} />,
  ];
  const featureCards = [
    {
      title: t('统一转发'),
      description: t('一个入口接入不同模型供应商，客户端配置更简单。'),
    },
    {
      title: t('兼容多协议'),
      description: t('支持 OpenAI、Claude、Gemini 等常见接口格式。'),
    },
    {
      title: t('密钥管理'),
      description: t('按用户和令牌管理权限，便于团队协作与隔离。'),
    },
    {
      title: t('用量统计'),
      description: t('集中查看调用量、额度和消费，运营更清楚。'),
    },
  ];
  const accessSteps = [
    {
      title: t('获取密钥'),
      description: t('登录控制台创建令牌，按需配置额度和权限。'),
    },
    {
      title: t('替换 Base URL'),
      description: t('把客户端的接口基址改成当前站点地址。'),
    },
    {
      title: t('选择模型'),
      description: t('继续使用熟悉的 SDK 和模型名称发起请求。'),
    },
  ];
  const endpointRows = [
    ['Chat Completions', '/v1/chat/completions'],
    ['Models', '/v1/models'],
    ['Images', '/v1/images/generations'],
  ];

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
          <div className='classic-home-hero w-full border-b border-semi-color-border relative overflow-hidden'>
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

            <div className='classic-home-shell classic-home-metrics'>
              <div>
                <strong>30+</strong>
                <span>{t('模型供应商')}</span>
              </div>
              <div>
                <strong>OpenAI</strong>
                <span>{t('兼容接口')}</span>
              </div>
              <div>
                <strong>API</strong>
                <span>{t('统一入口')}</span>
              </div>
            </div>
          </div>

          <section className='classic-home-section'>
            <div className='classic-home-shell'>
              <div className='classic-home-section-heading'>
                <Text>{t('为什么选择这个中转服务')}</Text>
                <h2>{t('把模型接入、密钥和用量放到一个清晰的工作台')}</h2>
              </div>
              <div className='classic-home-feature-grid'>
                {featureCards.map((item) => (
                  <article
                    className='classic-home-feature-card'
                    key={item.title}
                  >
                    <h3>{item.title}</h3>
                    <p>{item.description}</p>
                  </article>
                ))}
              </div>
            </div>
          </section>

          <section className='classic-home-section classic-home-section-muted'>
            <div className='classic-home-shell'>
              <div className='classic-home-section-heading'>
                <Text>{t('三步完成接入')}</Text>
                <h2>{t('保留熟悉的 SDK，只替换接入配置')}</h2>
              </div>
              <div className='classic-home-step-grid'>
                {accessSteps.map((step, index) => (
                  <article className='classic-home-step-card' key={step.title}>
                    <span>{String(index + 1).padStart(2, '0')}</span>
                    <h3>{step.title}</h3>
                    <p>{step.description}</p>
                  </article>
                ))}
              </div>
            </div>
          </section>

          <section className='classic-home-section'>
            <div className='classic-home-shell'>
              <div className='classic-home-section-heading classic-home-provider-heading'>
                <Text>{t('覆盖主流模型供应商')}</Text>
                <h2>{t('一个入口连接常用模型能力')}</h2>
                <p>{t('按需配置渠道和模型，让业务侧保持稳定的调用方式。')}</p>
              </div>
              <div className='classic-home-provider-grid'>
                {providerIcons.map((icon, index) => (
                  <div className='classic-home-provider-item' key={index}>
                    {icon}
                  </div>
                ))}
                <div className='classic-home-provider-item classic-home-provider-more'>
                  <Typography.Text>30+</Typography.Text>
                </div>
              </div>
            </div>
          </section>
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
