<script setup lang="ts">
import { ref } from 'vue'

const ws = ref<WebSocket | null>(null)
const isBroadcasting = ref(false)
const isListening = ref(false)

let mediaRecorder: MediaRecorder | null = null
let audioChunks: Blob[] = []
let audioContext: AudioContext | null = null
let source: MediaStreamAudioSourceNode | null = null

// Start broadcasting
const startBroadcasting = async () => {
  if (!navigator.mediaDevices.getUserMedia) {
    alert("Your browser does not support the MediaStream API")
    return
  }

  ws.value = new WebSocket('ws://localhost:8000/ws?mode=broadcast')

  ws.value.onopen = () => {
    console.log("Conected to the server as a broadcaster")
    isBroadcasting.value = true
  }

  const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
  mediaRecorder = new MediaRecorder(stream)

  mediaRecorder.ondataavailable = (e) => {
    ws.value?.send(e.data)
  }

  mediaRecorder.start(50) // Send audio in 50ms chunks
}
</script>