<script setup lang="ts">
import { ref } from 'vue'

const isBroadcasting = ref(false)
const isListening = ref(false)
const broadcastWs = ref<WebSocket | null>(null)
const listenerWs = ref<WebSocket | null>(null)

let mediaRecorder: MediaRecorder | null = null
let audioContext: AudioContext | null = null
let source: AudioBufferSourceNode | null = null

// Start broadcasting
const startBroadcasting = async () => {
  const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
  mediaRecorder = new MediaRecorder(stream, {
    mimeType: 'audio/webm;codecs=opus'
  })
  
  broadcastWs.value = new WebSocket('ws://localhost:8080/ws?mode=broadcast')
  
  broadcastWs.value.onopen = () => {
    console.log('Broadcaster WebSocket connected')
    mediaRecorder?.start(100)
    isBroadcasting.value = true
  }

  broadcastWs.value.onerror = (error) => {
    console.error('Broadcast WebSocket error:', error)
  }

  mediaRecorder.ondataavailable = (e) => {
    console.log('Broadcasting chunk size:', e.data.size)
    if (broadcastWs.value?.readyState === WebSocket.OPEN) {
      broadcastWs.value.send(e.data)
    }
  }
}

const stopBroadcasting = () => {
  mediaRecorder?.stop()
  broadcastWs.value?.close()
  isBroadcasting.value = false
}

// Start listening (Opt-in)
const startListening = () => {
  listenerWs.value = new WebSocket('ws://localhost:8080/ws?mode=listen')

  listenerWs.value.onopen = () => {
    console.log('Listener WebSocket connected')
    isListening.value = true
  }

  listenerWs.value.onmessage = async (event) => {
    console.log('Received audio data of type:', event.data.type)
    
    try {
      if (!audioContext) {
        audioContext = new AudioContext()
      }

      const buffer = await event.data.arrayBuffer()
      console.log('Buffer received, size:', buffer.byteLength)
      
      audioContext.decodeAudioData(
        buffer,
        (decoded) => {
          source = audioContext!.createBufferSource()
          source.buffer = decoded
          source.connect(audioContext!.destination)
          source.start()
        },
        (error) => {
          console.error('Error decoding audio:', error)
        }
      )
    } catch (error) {
      console.error('Error in audio processing:', error)
    }
  }
}

const stopListening = () => {
  if (source) source.stop()
  listenerWs.value?.close()
  isListening.value = false
}
</script>

<template>
  <div class="container text-center mt-5">
    <h1 class="mb-4 text-primary">🎤 EchoLive - Karaoke Broadcast</h1>

    <div class="d-flex justify-content-center gap-3">
      <!-- Start Broadcast Button -->
      <button 
        v-if="!isBroadcasting" 
        @click="startBroadcasting" 
        class="btn btn-success">
        🎙 Start Broadcasting
      </button>

      <!-- Stop Broadcast Button -->
      <button 
        v-if="isBroadcasting" 
        @click="stopBroadcasting" 
        class="btn btn-danger">
        ⛔ Stop Broadcasting
      </button>
    </div>

    <div class="mt-4 d-flex justify-content-center gap-3">
      <!-- Start Listening Button -->
      <button 
        v-if="!isListening" 
        @click="startListening" 
        class="btn btn-primary">
        🎧 Listen (Opt-in)
      </button>

      <!-- Stop Listening Button -->
      <button 
        v-if="isListening" 
        @click="stopListening" 
        class="btn btn-warning">
        🔇 Stop Listening
      </button>
    </div>
  </div>
</template>

<style scoped>
.container {
  max-width: 600px;
}
</style>