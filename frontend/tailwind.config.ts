import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        app: {
          bg: 'var(--color-bg)',
          surface: 'var(--color-surface)',
          'surface-muted': 'var(--color-surface-muted)',
          border: 'var(--color-border)',
          text: 'var(--color-text)',
          'text-muted': 'var(--color-text-muted)',
          'header-bg': 'var(--color-header-bg)',
          'header-text': 'var(--color-header-text)',
          brand: 'var(--color-brand)',
          primary: 'var(--color-primary)',
          'primary-hover': 'var(--color-primary-hover)',
          'primary-soft': 'var(--color-primary-soft)',
          focus: 'var(--color-focus)',
        },
        severity: {
          red: {
            icon: 'var(--color-severity-red-icon)',
            text: 'var(--color-severity-red-text)',
            bg: 'var(--color-severity-red-bg)',
            border: 'var(--color-severity-red-border)',
          },
          yellow: {
            icon: 'var(--color-severity-yellow-icon)',
            text: 'var(--color-severity-yellow-text)',
            bg: 'var(--color-severity-yellow-bg)',
            border: 'var(--color-severity-yellow-border)',
          },
          green: {
            icon: 'var(--color-severity-green-icon)',
            text: 'var(--color-severity-green-text)',
            bg: 'var(--color-severity-green-bg)',
            border: 'var(--color-severity-green-border)',
          },
        },
        status: {
          pending: {
            bg: 'var(--color-status-pending-bg)',
            text: 'var(--color-status-pending-text)',
          },
          ready: {
            bg: 'var(--color-status-ready-bg)',
            text: 'var(--color-status-ready-text)',
          },
          approved: {
            bg: 'var(--color-status-approved-bg)',
            text: 'var(--color-status-approved-text)',
          },
          notified: {
            bg: 'var(--color-status-notified-bg)',
            text: 'var(--color-status-notified-text)',
          },
          paid: {
            border: 'var(--color-status-paid-border)',
            text: 'var(--color-status-paid-text)',
          },
        },
      },
      borderRadius: {
        control: '6px',
        card: '10px',
      },
      keyframes: {
        highlight: {
          '0%': { backgroundColor: 'var(--color-primary-soft)' },
          '100%': { backgroundColor: 'transparent' },
        },
      },
      animation: {
        highlight: 'highlight 2s ease-out',
      },
    },
  },
  plugins: [],
} satisfies Config;
