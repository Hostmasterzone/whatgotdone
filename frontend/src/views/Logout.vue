<template>
  <div>
    <p>Signing out...</p>
  </div>
</template>

<script>
import getCsrfToken from '../controllers/CsrfToken.js';
import {logoutUserKit} from '../controllers/UserKit.js';

export default {
  name: 'Logout',
  created() {
    this.$store.commit('clearUsername');
    const url = `${process.env.VUE_APP_BACKEND_URL}/api/logout`;
    this.$http
      .post(url, {}, {headers: {'X-CSRF-Token': getCsrfToken()}})
      .then(() => {
        logoutUserKit();
        window.location.href = '/';
      })
      .finally(() => {
        // Logout can fail if CSRF goes out of state. In this case, still
        // delete the CSRF cookie.
        this.deleteCookie('csrf_base_v3');
      });
  },
  methods: {
    deleteCookie(name) {
      document.cookie = name + '=;expires=Thu, 01 Jan 1970 00:00:01 GMT';
    },
  },
};
</script>
