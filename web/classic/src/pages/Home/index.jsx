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
