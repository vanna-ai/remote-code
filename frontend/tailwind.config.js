/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        vanna: {
          navy: '#023d60',
          cream: '#e7e1cf',
          teal: '#15a8a8',
          orange: '#fe5d26',
          magenta: '#bf1363',
        },
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        gray: {
          50: '#f9fafb',
          100: '#f3f4f6',
          200: '#e5e7eb',
          300: '#d1d5db',
          400: '#9ca3af',
          500: '#6b7280',
          600: '#4b5563',
          700: '#374151',
          800: '#1f2937',
          900: '#111827',
        }
      },
      fontFamily: {
        sans: ['Space Grotesk', 'ui-sans-serif', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'Helvetica Neue', 'Arial', 'Noto Sans', 'sans-serif'],
        serif: ['Roboto Slab', 'ui-serif', 'Georgia', 'Cambria', 'Times New Roman', 'Times', 'serif'],
        mono: ['Disket Mono', 'Space Mono', 'JetBrains Mono', 'Monaco', 'Cascadia Code', 'Segoe UI Mono', 'Roboto Mono', 'monospace'],
      },
      boxShadow: {
        'vanna-subtle': '0 4px 15px rgba(21,168,168,0.2)',
        'vanna-card': '0 15px 40px rgba(15,23,42,0.08)',
        'vanna-feature': '0 25px 55px rgba(21,168,168,0.18)',
        'vanna-showcase': '0 30px 80px rgba(59,130,246,0.15)',
      },
    },
  },
  plugins: [],
}
