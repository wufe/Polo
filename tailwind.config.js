const Color = require('color');
const lighten = (color, val) => Color(color).lighten(val).rgba().string();

module.exports = {
  purge: ["./client/**/*.tsx", "./client/**/*.html"],
  darkMode: 'media', // or 'media' or 'class'
  theme: {
    extend: {
      minWidth: theme => ({
        ...theme('width'),
      }),
      "colors": theme => ({
        "nord-1": "#242932",
        "nord-2": "#20242c",
        "nord-3": "#1a1d23",
        "nord-4": "#16181d",
        "nord-5": "#0b0c0f",
        "nord4-alpha": Color("#D8DEE9").alpha(.5).string(),
        "nord6-alpha": Color("#ECEFF4").alpha(.5).string(),
        "nord14-alpha": Color("#A3BE8C").alpha(.5).string(),
        "nord11-alpha": Color("#BF616A").alpha(.5).string()
      }),
      maxWidth: theme => ({
        "1200": "1200px",
        "1500": "1500px"
      })
    },
    fontFamily: {
      'quicksand': ['Quicksand', 'sans-serif'],
      'mono': ['Courier New']
    }
  },
  variants: {
    extend: {},
  },
  plugins: [
    require('tailwind-nord'),
  ],
}
