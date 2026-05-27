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

// Broadsheet — After-hours newsprint tokens for Ant Design.
// Dark surfaces (page, cards, sidebars) with light cream form fields, so
// inputs read like a proof on the desk under a sodium lamp. No rounded
// corners, no shadows — steward's design principles applied to a dark deck.
const BG = '#1A1612'           // page bg — warm near-black
const PAPER = '#221D18'        // cards / sidebars / surfaces
const PAPER_DEEP = '#2C2620'   // raised / hover
const INK = '#F0E9DA'          // primary text — warm off-white
const INK_DIM = '#A39C8E'      // secondary
const INK_MUTE = '#8B8377'     // hint
const RULE = '#3B342D'         // hairline rules
const PRESS_RED = '#CA1625'    // primary / button fills
const PRESS_RED_SOFT = '#E26A55' // links / alerts — AA on dark
const PRESS_RED_INK = '#7E131E'  // pressed / active

// Light fields — cream paper sitting on the dark deck.
const FIELD_BG = '#FBF6EC'
const FIELD_BG_HOVER = '#F5EFE3'
const FIELD_TEXT = '#1A1A1A'
const FIELD_BORDER = '#3B342D'
const FIELD_PLACEHOLDER = '#6C6660'

const SANS = "-apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, 'Inter', system-ui, sans-serif"

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
    colorBgBase: BG,
    colorText: INK,
    colorTextSecondary: INK_DIM,
    colorTextTertiary: INK_MUTE,
    colorTextQuaternary: INK_MUTE,
    colorTextPlaceholder: INK_MUTE,
    colorTextDisabled: INK_MUTE,
    colorTextDescription: INK_DIM,
    colorBorder: RULE,
    colorBorderSecondary: RULE,
    colorBgLayout: BG,
    colorBgContainer: PAPER,
    colorBgElevated: PAPER,
    borderRadius: 0,
    borderRadiusLG: 0,
    borderRadiusSM: 0,
    borderRadiusXS: 0,
    fontFamily: SANS,
    fontFamilyCode: "ui-monospace, 'SF Mono', SFMono-Regular, 'JetBrains Mono', Menlo, monospace",
    wireframe: false
  },
  components: {
    Layout: {
      bodyBg: BG,
      headerBg: BG,
      lightSiderBg: BG,
      siderBg: BG
    },
    Button: {
      primaryShadow: 'none',
      defaultShadow: 'none',
      dangerShadow: 'none',
      fontWeight: 700,
      primaryColor: '#FFFFFF'
    },
    Card: {
      headerFontSize: 16,
      borderRadius: 0,
      borderRadiusLG: 0,
      borderRadiusSM: 0,
      borderRadiusXS: 0,
      colorBorderSecondary: RULE,
      colorBgContainer: PAPER
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
      colorBgElevated: PAPER
    },
    Modal: {
      colorBgElevated: PAPER,
      headerBg: PAPER
    },
    Timeline: {
      dotBg: BG
    },
    // Form widgets: cream paper field on the dark deck.
    Input: {
      colorBgContainer: FIELD_BG,
      colorBgContainerDisabled: FIELD_BG_HOVER,
      colorText: FIELD_TEXT,
      colorTextPlaceholder: FIELD_PLACEHOLDER,
      colorBorder: FIELD_BORDER,
      activeBorderColor: PRESS_RED,
      hoverBorderColor: PRESS_RED_INK,
      activeShadow: 'none'
    },
    InputNumber: {
      colorBgContainer: FIELD_BG,
      colorText: FIELD_TEXT,
      colorTextPlaceholder: FIELD_PLACEHOLDER,
      colorBorder: FIELD_BORDER,
      activeBorderColor: PRESS_RED,
      hoverBorderColor: PRESS_RED_INK,
      activeShadow: 'none'
    },
    Select: {
      colorBgContainer: FIELD_BG,
      colorBgElevated: FIELD_BG,
      colorText: FIELD_TEXT,
      colorTextPlaceholder: FIELD_PLACEHOLDER,
      colorBorder: FIELD_BORDER,
      optionSelectedBg: 'rgba(202, 22, 37, 0.12)',
      optionSelectedColor: FIELD_TEXT,
      optionActiveBg: 'rgba(26, 22, 18, 0.06)'
    },
    DatePicker: {
      colorBgContainer: FIELD_BG,
      colorBgElevated: FIELD_BG,
      colorText: FIELD_TEXT,
      colorTextPlaceholder: FIELD_PLACEHOLDER,
      colorBorder: FIELD_BORDER,
      activeBorderColor: PRESS_RED,
      cellHoverBg: 'rgba(202, 22, 37, 0.10)',
      cellActiveWithRangeBg: 'rgba(202, 22, 37, 0.16)'
    },
    Cascader: {
      colorBgContainer: FIELD_BG,
      colorText: FIELD_TEXT
    },
    Checkbox: {
      colorBgContainer: FIELD_BG,
      colorBorder: FIELD_BORDER
    },
    Radio: {
      colorBgContainer: FIELD_BG,
      colorBorder: FIELD_BORDER
    },
    Form: {
      labelColor: INK_DIM,
      labelFontSize: 12
    },
    Typography: {
      titleMarginBottom: '0.5em',
      titleMarginTop: '1em',
      fontWeightStrong: 700
    },
    Tabs: {
      itemActiveColor: PRESS_RED_SOFT,
      itemSelectedColor: PRESS_RED_SOFT,
      itemHoverColor: PRESS_RED,
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
    },
    Steps: {
      colorTextDescription: INK_DIM,
      // Pending steps need to stay legible on the dark deck
      colorTextLabel: INK_DIM
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
