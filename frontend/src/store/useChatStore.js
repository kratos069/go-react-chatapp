import { create } from "zustand";
import toast from "react-hot-toast";
import { axiosInstance } from "../lib/axios";
import { useAuthStore } from "./useAuthStore";

export const useChatStore = create((set, get) => ({
  messages: [],
  users: [],
  selectedUser: null,
  isUsersLoading: false,
  isMessagesLoading: false,

  getUsers: async () => {
    set({ isUsersLoading: true });
    try {
      const res = await axiosInstance.get("/messages/users");
      set({ users: res.data.users });
    } catch (error) {
      toast.error(error.response?.data?.error || "Error fetching users");
    } finally {
      set({ isUsersLoading: false });
    }
  },

  getMessages: async (userId) => {
    set({ isMessagesLoading: true });
    try {
      const res = await axiosInstance.get(`/messages/${userId}`);
      set({ messages: res.data.messages });
    } catch (error) {
      toast.error(error.response?.data?.message || "Error fetching messages");
    } finally {
      set({ isMessagesLoading: false });
    }
  },

  sendMessage: async (messageData) => {
    const { selectedUser, messages } = get();
    if (!selectedUser) {
      toast.error("No user selected");
      return;
    }

    try {
      const res = await axiosInstance.post(
        `/messages/send/${selectedUser.id}`,
        messageData
      );
      set({ messages: [...messages, res.data.message] });
    } catch (error) {
      toast.error(error.response?.data?.error || "Failed to send message");
    }
  },

  subscribeToMessages: () => {
    const { selectedUser } = get();
    const { socket, connectSocket } = useAuthStore.getState(); // Access connectSocket from useAuthStore

    if (!selectedUser) {
      console.error("No selected user to subscribe");
      return;
    }

    const onSocketMessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (
          data.event === "newMessage" &&
          data.message.senderId === selectedUser.id
        ) {
          set((state) => ({ messages: [...state.messages, data.message] }));
        }
      } catch (error) {
        console.error("WebSocket message error:", error);
      }
    };

    if (!socket || socket.readyState !== WebSocket.OPEN) {
      connectSocket(() => {
        socket.onmessage = onSocketMessage;
      });
    } else {
      socket.onmessage = onSocketMessage;
    }
  },

  unsubscribeFromMessages: () => {
    const { socket } = useAuthStore.getState(); // Access socket from useAuthStore
    if (socket) {
      socket.onmessage = null;
      console.log("Unsubscribed from WebSocket messages");
    }
  },

  setSelectedUser: (selectedUser) => set({ selectedUser }),
}));
