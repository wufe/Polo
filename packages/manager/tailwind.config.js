const path = require('path');
const Color = require('color');
const lighten = (color, val) => Color(color).lighten(val).rgba().string();

module.exports = {
  mode: 'jit',
  purge: [
    path.resolve(__dirname, "./src/**/*.tsx"),
    path.resolve(__dirname, "./src/**/*.html"),
  ],
  darkMode: 'media', // or 'media' or 'class'
  theme: {
    extend: {
      minWidth: theme => ({
        ...theme('width'),
      }),
      "colors": theme => ({
        'black-alpha10': Color('#000').alpha(.1).string(),
        'white-alpha10': Color('#FFF').alpha(.1).string(),
        "nord-1": "#242932",
        "nord-2": "#20242c",
        "nord-3": "#1a1d23",
        "nord-4": "#16181d",
        "nord-5": "#0b0c0f",
        "nord0-alpha10": Color("#2E3440").alpha(.1).string(),
        "nord0-alpha20": Color("#2E3440").alpha(.2).string(),
        "nord0-alpha30": Color("#2E3440").alpha(.3).string(),
        "nord0-alpha40": Color("#2E3440").alpha(.4).string(),
        "nord0-alpha50": Color("#2E3440").alpha(.5).string(),
        "nord4-alpha10": Color("#D8DEE9").alpha(.1).string(),
        "nord4-alpha20": Color("#D8DEE9").alpha(.2).string(),
        "nord4-alpha30": Color("#D8DEE9").alpha(.3).string(),
        "nord4-alpha40": Color("#D8DEE9").alpha(.4).string(),
        "nord4-alpha50": Color("#D8DEE9").alpha(.5).string(),
        "nord4-alpha": Color("#D8DEE9").alpha(.5).string(),
        "nord6-alpha10": Color("#ECEFF4").alpha(.1).string(),
        "nord6-alpha20": Color("#ECEFF4").alpha(.2).string(),
        "nord6-alpha30": Color("#ECEFF4").alpha(.3).string(),
        "nord6-alpha40": Color("#ECEFF4").alpha(.4).string(),
        "nord6-alpha50": Color("#ECEFF4").alpha(.5).string(),
        "nord6-alpha": Color("#ECEFF4").alpha(.5).string(),
        "nord14-alpha50": Color("#A3BE8C").alpha(.5).string(),
        "nord14-alpha": Color("#A3BE8C").alpha(.5).string(),
        "nord11-alpha50": Color("#BF616A").alpha(.5).string(),
        "nord11-alpha": Color("#BF616A").alpha(.5).string(),
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
