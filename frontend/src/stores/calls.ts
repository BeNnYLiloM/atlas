import { defineStore } from 'pinia'
import { ref } from 'vue'
import { Room, RoomEvent, Track } from 'livekit-client'
import type { RemoteParticipant, RemoteTrackPublication, RemoteTrack } from 'livekit-client'
import apiClient from '@/api/client'

// Скрытый div для аудио элементов (LiveKit требует attach треков к DOM)
let audioContainer: HTMLDivElement | null = null
function getAudioContainer(): HTMLDivElement {
  if (!audioContainer) {
    audioContainer = document.createElement('div')
    audioContainer.style.display = 'none'
    audioContainer.id = 'livekit-audio-container'
    document.body.appendChild(audioContainer)
  }
  return audioContainer
}

interface CallToken {
  token: string
  room_name: string
  url: string
}

export const useCallsStore = defineStore('calls', () => {
  const room = ref<Room | null>(null)
  const isInCall = ref(false)
  const isMuted = ref(false)
  const isCameraOff = ref(false)
  const participants = ref<string[]>([])
  const currentRoomName = ref<string | null>(null)
  const currentChannelId = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  function isInChannel(channelId: string): boolean {
    return isInCall.value && currentChannelId.value === channelId
  }

  async function toggleVoiceChannel(channelId: string) {
    if (isInChannel(channelId)) {
      await leaveCall()
      return
    }
    await joinCall(channelId)
  }

  async function joinCall(channelId: string) {
    if (isInCall.value) await leaveCall()

    loading.value = true
    error.value = null

    try {
      const res = await apiClient.post<{ data: CallToken }>('/calls/join', {
        channel_id: channelId,
      })

      const { token, url, room_name } = res.data.data
      currentRoomName.value = room_name
      currentChannelId.value = channelId

      const newRoom = new Room({
        audioCaptureDefaults: { echoCancellation: true, noiseSuppression: true },
      })
      room.value = newRoom

      // Attach аудио треков удалённых участников для воспроизведения
      newRoom.on(RoomEvent.TrackSubscribed, (track: RemoteTrack, _pub: RemoteTrackPublication, _participant: RemoteParticipant) => {
        if (track.kind === Track.Kind.Audio) {
          const el = track.attach()
          getAudioContainer().appendChild(el)
        }
      })

      newRoom.on(RoomEvent.TrackUnsubscribed, (track: RemoteTrack) => {
        if (track.kind === Track.Kind.Audio) {
          track.detach().forEach((el: HTMLMediaElement) => el.remove())
        }
      })

      newRoom.on(RoomEvent.ParticipantConnected, () => updateParticipants())
      newRoom.on(RoomEvent.ParticipantDisconnected, () => updateParticipants())
      newRoom.on(RoomEvent.Disconnected, () => {
        isInCall.value = false
        room.value = null
        currentChannelId.value = null
        if (audioContainer) audioContainer.innerHTML = ''
      })

      await newRoom.connect(url, token)

      // Attach аудио от участников уже находящихся в комнате
      newRoom.participants.forEach(participant => {
        (participant as RemoteParticipant).tracks.forEach((pub: RemoteTrackPublication) => {
          if (pub.track && pub.track.kind === Track.Kind.Audio) {
            const el = (pub.track as RemoteTrack).attach()
            getAudioContainer().appendChild(el)
          }
        })
      })

      // Включаем микрофон
      try {
        await newRoom.localParticipant.setMicrophoneEnabled(true)
      } catch {
        console.warn('[Calls] No microphone, joining as listener')
      }

      isInCall.value = true
      updateParticipants()
    } catch (e) {
      error.value = 'Не удалось подключиться к звонку'
      console.error('[Calls] Join error:', e)
    } finally {
      loading.value = false
    }
  }

  async function leaveCall() {
    if (room.value) {
      await room.value.disconnect()
      room.value = null
    }
    isInCall.value = false
    currentRoomName.value = null
    currentChannelId.value = null
    participants.value = []
    if (audioContainer) audioContainer.innerHTML = ''
  }

  async function toggleMute() {
    if (!room.value) return
    const enabled = !isMuted.value
    await room.value.localParticipant.setMicrophoneEnabled(!enabled)
    isMuted.value = enabled
  }

  async function toggleCamera() {
    if (!room.value) return
    const enabled = !isCameraOff.value
    await room.value.localParticipant.setCameraEnabled(!enabled)
    isCameraOff.value = enabled
  }

  function updateParticipants() {
    if (!room.value) return
    // v1 SDK: room.participants — Map удалённых участников
    participants.value = Array.from(room.value.participants.values())
      .map(p => (p as RemoteParticipant).name || (p as RemoteParticipant).identity)
  }

  function $reset() {
    room.value = null
    isInCall.value = false
    isMuted.value = false
    isCameraOff.value = false
    participants.value = []
    currentRoomName.value = null
    currentChannelId.value = null
    loading.value = false
    error.value = null
  }

  return {
    room,
    isInCall,
    isMuted,
    isCameraOff,
    participants,
    currentRoomName,
    currentChannelId,
    loading,
    error,
    isInChannel,
    joinCall,
    toggleVoiceChannel,
    leaveCall,
    toggleMute,
    toggleCamera,
    $reset,
  }
})
