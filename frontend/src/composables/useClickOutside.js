import { onMounted, onBeforeUnmount } from 'vue'

export function onClickOutside(elementRef, callback) {
  function handler(e) {
    if (elementRef.value && !elementRef.value.contains(e.target)) {
      callback(e)
    }
  }
  onMounted(() => document.addEventListener('mousedown', handler))
  onBeforeUnmount(() => document.removeEventListener('mousedown', handler))
}
