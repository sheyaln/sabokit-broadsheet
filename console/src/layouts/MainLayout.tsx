import { Layout } from 'antd'
import { ReactNode } from 'react'

const { Content } = Layout

interface MainLayoutProps {
  children: ReactNode
}

/*
 * MainLayout — the unauthenticated chrome. Used by sign-in, accept-invitation,
 * setup wizard, and the workspace picker dashboard. Newsprint cream paper
 * with a broadsheet masthead at the top, a date-line beneath, and a hand-set
 * "Vol. / No." eyebrow. The aesthetic nods to mid-century union papers
 * without putting on a costume.
 */
export function MainLayout({ children }: MainLayoutProps) {
  const today = new Date().toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })

  return (
    <Layout
      style={{
        minHeight: '100vh',
        backgroundColor: '#1a1612',
        backgroundImage:
          'radial-gradient(rgba(240, 233, 218, 0.04) 1px, transparent 1px)',
        backgroundSize: '4px 4px'
      }}
    >
      <div style={{ maxWidth: 980, margin: '0 auto', padding: '40px 24px 12px', width: '100%' }}>
        <div
          style={{
            borderTop: '3px solid #5a4f43',
            borderBottom: '1px solid #5a4f43',
            padding: '14px 0 10px',
            display: 'flex',
            alignItems: 'baseline',
            justifyContent: 'space-between',
            gap: 24
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10
            }}
          >
            <img
              src="/console/broadsheet-icon.png"
              alt=""
              width={44}
              height={44}
              style={{ display: 'block' }}
            />
            <div
              style={{
                fontFamily: "'IBM Plex Sans', system-ui, sans-serif",
                fontSize: 10,
                fontWeight: 600,
                letterSpacing: '0.18em',
                textTransform: 'uppercase',
                color: '#a39c8e',
                lineHeight: 1.3
              }}
            >
              Vol. {window.VERSION || '1.0'}
              <br />
              No. 1
            </div>
          </div>
          <div
            style={{
              fontFamily: "'Fraunces', 'IBM Plex Serif', Georgia, serif",
              fontWeight: 800,
              fontSize: 56,
              color: '#f0e9da',
              letterSpacing: '-0.03em',
              lineHeight: 1,
              textAlign: 'center'
            }}
          >
            Broad<span style={{ color: '#ca1625' }}>side</span>
          </div>
          <div
            style={{
              fontFamily: "'IBM Plex Sans', system-ui, sans-serif",
              fontSize: 10,
              fontWeight: 600,
              letterSpacing: '0.18em',
              textTransform: 'uppercase',
              color: '#a39c8e',
              textAlign: 'right'
            }}
          >
            {today}
          </div>
        </div>
        <div
          style={{
            height: 8,
            borderBottom: '1px solid #3b342d'
          }}
        />
      </div>
      <Content style={{ padding: '24px' }}>{children}</Content>
    </Layout>
  )
}

interface MainLayoutSidebarProps {
  children: ReactNode
  title: string
  extra: ReactNode
}

export function MainLayoutSidebar({ children, title, extra }: MainLayoutSidebarProps) {
  return (
    <div
      className="fixed right-0 top-0 bottom-0 w-[400px] p-6 overflow-y-auto"
      style={{
        backgroundColor: '#221d18',
        borderLeft: '1px solid #3b342d'
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '20px',
          paddingBottom: '12px',
          borderBottom: '1px solid #3b342d'
        }}
      >
        <h3
          style={{
            margin: 0,
            fontFamily: "'Fraunces', 'IBM Plex Serif', Georgia, serif",
            fontWeight: 700,
            letterSpacing: '-0.015em'
          }}
        >
          {title}
        </h3>
        {extra}
      </div>
      {children}
    </div>
  )
}
