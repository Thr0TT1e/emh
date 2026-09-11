import { definePreset } from '@primeuix/themes';
import Aura from '@primeuix/themes/aura';

const emhPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#fdf3f4',
      100: '#f9dde0',
      200: '#f2b8bd',
      300: '#e68a92',
      400: '#d4545f',
      500: '#bd2f3c',
      600: '#a31621',
      700: '#8a121c',
      800: '#700f17',
      900: '#5a0c13',
      950: '#43090e',
    },
  },
});

export default {
  preset: emhPreset,
  options: {
    darkModeSelector: '.p-dark',
  },
};
