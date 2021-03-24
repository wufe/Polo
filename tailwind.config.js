const Color = require('color');
const lighten = (color, val) => Color(color).lighten(val).rgba().string();

module.exports = {
  purge: ["./client/**/*.tsx", "./client/**/*.html"],
  darkMode: 'media', // or 'media' or 'class'
  theme: {
    extend: {
      "colors": theme => ({
        "nord-1": "#242932",
        "nord-2": "#20242c",
        "nord-3": "#1a1d23",
        "nord-4": "#16181d",
        "nord-5": "#0b0c0f",
        "nord14-alpha": Color("#A3BE8C").alpha(.5).string(),
        "nord11-alpha": Color("#BF616A").alpha(.5).string()
      })
    },
    fontFamily: {
      'quicksand': ['Quicksand', 'sans-serif'],
      'mono': ['Courier New']
    },
    minWidth: {
      "0": "0",
      "16": "4rem"
    }
  },
  variants: {
    extend: {},
  },
  plugins: [
    require('tailwind-nord'),
  ],
}
