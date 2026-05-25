import { ConfigProvider, App as AntApp, ThemeConfig, theme as antdTheme } from 'antd'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { I18nProvider } from '@lingui/react'
import { router } from './router'
import { AuthProvider } from './contexts/AuthContext'
import { LocaleProvider, useLocale, i18n } from './contexts/LocaleContext'
import { initializeAnalytics } from './utils/analytics-config'
import enUS from 'antd/locale/en_US'
import frFR from 'antd/locale/fr_FR'
import esES from 'antd/locale/es_ES'
import deDE from 'antd/locale/de_DE'
import caES from 'antd/locale/ca_ES'
import type { Locale as AntdLocale } from 'antd/es/locale'
import type { Locale } from './i18n'

const antdLocales: Record<Locale, AntdLocale> = {
  en: enUS,
  fr: frFR,
  es: esES,
  de: deDE,
  ca: caES,
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1
    }
  }
})

// Broadsheet — Newsprint (after-hours) tokens for Ant Design.
// Dark variant: warm darks, brick-red re-tuned for AA on near-black.
// See console/THEMES.md for the design notes and alternative directions.
const PAPER = '#1A1612'         // page bg — warm near-black
const PAPER_BRIGHT = '#221D18'  // surfaces / cards
const PAPER_DEEP = '#2C2620'    // raised / hover
const INK = '#F0E9DA'           // primary text — warm off-white
const INK_MUTED = '#A39C8E'     // secondary text
const INK_HINT = '#8B8377'      // placeholder / empty-state hint — readable on PAPER_BRIGHT
const RULE = '#3B342D'          // hairline rules
const PRESS_RED = '#CA1625'     // logo body — vivid pure red
const PRESS_RED_DEEP = '#A8121F'    // pressed
const PRESS_RED_SOFT = '#F61D28'    // logo peak — used where AA on dark body text matters (links, alerts)

const SANS = "'IBM Plex Sans', system-ui, -apple-system, 'Segoe UI', sans-serif"

const theme: ThemeConfig = {
  algorithm: antdTheme.darkAlgorithm,
  token: {
    colorPrimary: PRESS_RED,
    colorLink: PRESS_RED_SOFT,
    colorLinkHover: PRESS_RED,
    colorInfo: '#6AA2C4',
    colorSuccess: '#7BB364',
    colorWarning: '#D8A35A',
    colorError: PRESS_RED_SOFT,
    colorTextBase: INK,
    colorBgBase: PAPER,
    colorText: INK,
    colorTextSecondary: INK_MUTED,
    colorTextTertiary: INK_HINT,
    colorTextQuaternary: INK_HINT,
    colorTextPlaceholder: INK_HINT,
    colorTextDisabled: INK_HINT,
    colorTextDescription: INK_MUTED,
    colorBorder: RULE,
    colorBorderSecondary: RULE,
    colorBgLayout: PAPER,
    colorBgContainer: PAPER_BRIGHT,
    colorBgElevated: PAPER_BRIGHT,
    borderRadius: 2,
    borderRadiusLG: 3,
    borderRadiusSM: 2,
    borderRadiusXS: 2,
    fontFamily: SANS,
    fontFamilyCode: "'IBM Plex Mono', ui-monospace, Menlo, monospace",
    wireframe: false
  },
  components: {
    Layout: {
      bodyBg: PAPER,
      headerBg: PAPER,
      lightSiderBg: PAPER,
      siderBg: PAPER
    },
    Button: {
      primaryShadow: 'none',
      defaultShadow: 'none',
      dangerShadow: 'none',
      fontWeight: 500,
      // Primary buttons: white text on logo-red hits 6.2:1 (AA / AAA for large)
      primaryColor: '#FFFFFF'
    },
    Card: {
      headerFontSize: 16,
      borderRadius: 2,
      borderRadiusLG: 3,
      borderRadiusSM: 2,
      borderRadiusXS: 2,
      colorBorderSecondary: RULE,
      colorBgContainer: PAPER_BRIGHT
    },
    Table: {
      headerBg: 'transparent',
      fontSize: 12,
      colorTextHeading: INK,
      colorBgContainer: 'transparent',
      rowHoverBg: 'rgba(240, 233, 218, 0.04)',
      headerSplitColor: RULE,
      borderColor: RULE
    },
    Menu: {
      itemBg: 'transparent',
      subMenuItemBg: 'transparent',
      itemSelectedBg: 'rgba(226, 106, 85, 0.14)',
      itemSelectedColor: PRESS_RED_SOFT,
      itemHoverBg: 'rgba(240, 233, 218, 0.06)',
      itemHoverColor: INK,
      itemColor: INK,
      itemHeight: 38,
      iconSize: 14
    },
    Drawer: {
      colorBgElevated: PAPER_BRIGHT
    },
    Modal: {
      colorBgElevated: PAPER_BRIGHT,
      headerBg: PAPER_BRIGHT
    },
    Timeline: {
      dotBg: PAPER
    },
    Input: {
      colorBgContainer: PAPER_DEEP,
      activeBorderColor: PRESS_RED,
      hoverBorderColor: INK_MUTED
    },
    Select: {
      colorBgContainer: PAPER_DEEP
    },
    Typography: {
      titleMarginBottom: '0.5em',
      titleMarginTop: '1em',
      fontWeightStrong: 600
    },
    Tabs: {
      itemActiveColor: PRESS_RED,
      itemSelectedColor: PRESS_RED,
      itemHoverColor: PRESS_RED_SOFT,
      inkBarColor: PRESS_RED
    },
    Tag: {
      defaultBg: PAPER_DEEP,
      defaultColor: INK
    },
    Divider: {
      colorSplit: RULE
    },
    Tooltip: {
      colorBgSpotlight: PAPER_DEEP,
      colorTextLightSolid: INK
    }
  }
}

// Initialize analytics service
initializeAnalytics()

// Inner component that uses LocaleContext
function AppContent() {
  const { locale } = useLocale()

  return (
    // key={locale} forces I18nProvider and all children to remount when locale changes,
    // ensuring all components re-render with the new translations
    <I18nProvider i18n={i18n} key={locale}>
      <ConfigProvider theme={theme} locale={antdLocales[locale]}>
        <AntApp>
          <RouterProvider router={router} />
        </AntApp>
      </ConfigProvider>
    </I18nProvider>
  )
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <LocaleProvider>
          <AppContent />
        </LocaleProvider>
      </AuthProvider>
    </QueryClientProvider>
  )
}

export default App
