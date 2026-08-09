// OpsFlow Service Worker — Shell caching only for PWA installability.
// Full offline support is not an MVP goal (design.md § 14).

const CACHE_NAME = 'opsflow-shell-v1';
const SHELL_ASSETS = [
  '/',
  '/index.html',
  '/src/main.js',
  '/src/style.css',
  '/manifest.json',
  '/icons/favicon.svg',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(SHELL_ASSETS))
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys
          .filter((key) => key !== CACHE_NAME)
          .map((key) => caches.delete(key))
      )
    )
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  // Network-first for API calls, cache-first for shell assets.
  if (event.request.url.includes('/api/')) {
    event.respondWith(fetch(event.request));
    return;
  }

  event.respondWith(
    caches.match(event.request).then((cached) => cached || fetch(event.request))
  );
});
