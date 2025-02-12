import { create } from "zustand";
import { axiosInstance } from "../lib/axios.js";
import toast from "react-hot-toast";

const BASE_URL =
  import.meta.env.MODE === "development"
    ? "ws://localhost:3000/api/ws"
    : "/api/ws";

export const useAuthStore = create((set, get) => ({
  authUser: null,
  isSigningUp: false,
  isLoggingIn: false,
  isUpdatingProfile: false,
  isCheckingAuth: true,
  onlineUsers: [],
  socket: null,

  checkAuth: async () => {
    try {
      const res = await axiosInstance.get("/auth/check");
      set({ authUser: res.data.user });
      get().connectSocket();
    } catch (error) {
      console.log("Error in checkAuth:", error);
      set({ authUser: null });
    } finally {
      set({ isCheckingAuth: false });
    }
  },

  signup: async (data) => {
    set({ isSigningUp: true });
    try {
      const res = await axiosInstance.post("/auth/signup", data);
      set({ authUser: res.data.user });
      toast.success("Account created successfully");
      get().connectSocket();
    } catch (error) {
      toast.error(error.response?.data?.error || "Signup failed");
    } finally {
      set({ isSigningUp: false });
    }
  },

  login: async (data) => {
    set({ isLoggingIn: true });
    try {
      const res = await axiosInstance.post("/auth/login", data);
      set({ authUser: res.data.user });
      toast.success("Logged in successfully");
      get().connectSocket();
    } catch (error) {
      toast.error(error.response?.data?.error || "Login failed");
    } finally {
      set({ isLoggingIn: false });
    }
  },

  logout: async () => {
    try {
      await axiosInstance.post("/auth/logout");
      set({ authUser: null });
      toast.success("Logged out successfully");
      get().disconnectSocket();
    } catch (error) {
      toast.error(error.response?.data?.message || "Logout failed");
    }
  },

  updateProfile: async (data) => {
    set({ isUpdatingProfile: true });
    try {
      const res = await axiosInstance.put("/user/update-profile", data);
      set({ authUser: res.data.user });
      toast.success("Profile updated successfully");
    } catch (error) {
      console.log("Error in update profile:", error);
      toast.error(error.response?.data?.error || "Profile update failed");
    } finally {
      set({ isUpdatingProfile: false });
    }
  },

  connectSocket: (callback) => {
    const { authUser } = get();

    if (!authUser) {
      console.error("User is not authenticated");
      return;
    }

    if (get().socket?.readyState === WebSocket.OPEN) {
      console.log("WebSocket is already connected");
      callback?.(); // Call the callback immediately if already connected
      return;
    }

    const socket = new WebSocket(BASE_URL);

    socket.onopen = () => {
      console.log("WebSocket connected");
      set({ socket });
      socket.send(JSON.stringify({ type: "JOIN", userId: authUser.id }));
      callback?.(); // Trigger the callback after the socket is open
    };

    socket.onclose = () => {
      console.warn("WebSocket disconnected. Attempting to reconnect...");
      setTimeout(() => get().connectSocket(callback), 3000);
    };

    socket.onerror = (err) => {
      console.error("WebSocket connection error:", err);
    };

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);

      if (message.event === "getOnlineUsers") {
        set({ onlineUsers: message.onlineUsers }); // Update online users
      }
    };
  },

  disconnectSocket: () => {
    const { socket } = get();

    if (socket) {
      socket.close();
      set({ socket: null, onlineUsers: [] });
      console.log("WebSocket disconnected");
    }
  },
}));
