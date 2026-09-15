import { useMemo } from 'react';
import { ConfigProvider, Layout } from 'antd';

import { useTheme } from '@/hooks/useTheme';
import AppSidebar from '@/layouts/AppSidebar';
import { ensurePgAdminI18n } from '@/pg-ui/i18n/admin-pages-bridge';
import PgAdminsContent from '@/pg-ui/pages/_dashboard.admins';

ensurePgAdminI18n();

export default function AdminsPage() {
  const { isDark, isUltra, antdThemeConfig } = useTheme();

  const pageClass = useMemo(() => {
    const classes = ['admins-page'];
    if (isDark) classes.push('is-dark');
    if (isUltra) classes.push('is-ultra');
    return classes.join(' ');
  }, [isDark, isUltra]);

  return (
    <ConfigProvider theme={antdThemeConfig}>
      <Layout className={pageClass}>
        <AppSidebar />
        <Layout className="content-shell">
          <Layout.Content id="content-layout" className="content-area">
            <PgAdminsContent />
          </Layout.Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
}
