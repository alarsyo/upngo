import Vue from 'vue';
import Vuetify from 'vuetify/lib';
import colors from 'vuetify/es5/util/colors';
import 'vuetify/src/stylus/app.styl';

Vue.use(Vuetify, {
  iconfont: 'md',
  theme: {
    primary: colors.pink.darken1,
    secondary: colors.pink.lighten3,
    accent: colors.green.accent1,
  },
});
