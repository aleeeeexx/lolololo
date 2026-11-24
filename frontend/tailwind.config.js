/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        background: 'var(--background)',
        surface: 'var(--surface)',
        'surface-deep': 'var(--surface-deep)',
        foreground: 'var(--foreground)',
        'foreground-secondary': 'var(--foreground-secondary)',
        muted: 'var(--muted)',
        // DEFAULT / named keys are the fluorescent brand tokens.
        // Numeric scale is ink-olive so existing text-primary-600 links
        // and focus rings stay olive, never fluorescent green.
        primary: {
          DEFAULT: 'var(--primary)',
          foreground: 'var(--primary-foreground)',
          hover: 'var(--primary-hover)',
          active: 'var(--primary-active)',
          50: '#F4F7E3',
          100: '#E8F0B8',
          200: '#D4E07A',
          300: '#C5D46A',
          400: '#A8B84A',
          500: '#4D5B00',
          600: '#4D5B00',
          700: '#3D4900',
          800: '#2A3200',
          900: '#14150B',
          950: '#0C0D06'
        },
        accent: {
          ink: 'var(--accent-ink)',
          hover: 'var(--accent-hover)',
          soft: 'var(--accent-soft)',
          50: '#F4F7E3',
          100: '#E8F0B8',
          200: '#D4E07A',
          300: '#C5D46A',
          400: '#A8B84A',
          500: '#4D5B00',
          600: '#4D5B00',
          700: '#3D4900',
          800: '#2A3200',
          900: '#14150B',
          950: '#0C0D06'
        },
        dark: {
          50: '#F0F0FA',
          100: '#E4E4EE',
          200: '#CCCCCC',
          300: '#CCCCCC',
          400: '#9B9BA6',
          500: '#9B9BA6',
          600: '#252628',
          700: '#252628',
          800: '#161618',
          900: '#0A0A0C',
          950: '#000000'
        },
        gray: {
          50: '#F6F7F2',
          100: '#F0F1EB',
          200: '#E5E8DE',
          300: '#D3D6CC',
          400: '#A8AB9F',
          500: '#73766B',
          600: '#474B40',
          700: '#3A3D34',
          800: '#2A2C26',
          900: '#15160F',
          950: '#14150B'
        }
      },
      borderColor: {
        DEFAULT: 'var(--border)',
        strong: 'var(--border-strong)'
      },
      borderRadius: {
        none: '0',
        sm: 'var(--radius-sm)',
        DEFAULT: 'var(--radius)',
        md: 'var(--radius)',
        lg: 'var(--radius)',
        xl: 'var(--radius)',
        '2xl': 'var(--radius)',
        '3xl': 'var(--radius)',
        '4xl': 'var(--radius)'
      },
      fontFamily: {
        sans: [
          'Inter',
          '"Noto Sans SC"',
          '"PingFang SC"',
          '"Microsoft YaHei"',
          'system-ui',
          'sans-serif'
        ],
        display: [
          'Inter',
          '"Noto Sans SC"',
          '"PingFang SC"',
          '"Microsoft YaHei"',
          'system-ui',
          'sans-serif'
        ],
        mono: [
          '"JetBrains Mono"',
          'ui-monospace',
          'SFMono-Regular',
          'Menlo',
          'Monaco',
          'Consolas',
          'monospace'
        ]
      },
      boxShadow: {
        glass: 'none',
        'glass-sm': 'none',
        glow: 'none',
        'glow-lg': 'none',
        card: 'none',
        'card-hover': 'none',
        'inner-glow': 'none',
        sm: 'none',
        DEFAULT: 'none',
        md: 'none',
        lg: 'none',
        xl: 'none',
        '2xl': 'none'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': 'none',
        'gradient-dark': 'none',
        'gradient-glass': 'none',
        'mesh-gradient': 'none'
      },
      maxWidth: {
        content: '1240px'
      },
      transitionDuration: {
        DEFAULT: '150ms'
      },
      animation: {
        'fade-in': 'fadeIn 0.15s ease-out',
        'slide-up': 'slideUp 0.15s ease-out',
        'slide-down': 'slideDown 0.15s ease-out',
        'slide-in-right': 'slideInRight 0.15s ease-out',
        'scale-in': 'scaleIn 0.15s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'none'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideDown: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideInRight: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        scaleIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        }
      },
      backdropBlur: {
        xs: '2px'
      }
    }
  },
  plugins: []
}
