import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || "http://localhost:8080/api/v1";

/**
 * Auth Service - Handles authentication operations
 */
class AuthService {
  /**
   * Login user with email/username and password
   * @param {string} identifier - Username or email
   * @param {string} password - User password
   * @returns {Promise<{accessToken: string, refreshToken: string, expiresIn: number}>}
   */
  async login(identifier, password) {
    try {
      const response = await axios.post(`${API_BASE_URL}/auth/login`, {
        identifier: identifier.trim(),
        password,
      });

      const { accessToken, refreshToken, expiresIn } = response.data;

      if (!accessToken || !refreshToken) {
        throw new Error('Invalid response: missing tokens');
      }

      // Store tokens
      this.setAccessToken(accessToken);
      this.setRefreshToken(refreshToken);

      return {
        accessToken,
        refreshToken,
        expiresIn,
      };
    } catch (error) {
      const errorMessage = error.response?.data?.error || error.message || 'Login failed';
      throw new Error(errorMessage);
    }
  }

  /**
   * Refresh access token using refresh token
   * @returns {Promise<{accessToken: string, refreshToken: string, expiresIn: number}>}
   */
  async refreshToken() {
    try {
      const refreshToken = this.getRefreshToken();
      
      if (!refreshToken) {
        throw new Error('No refresh token available');
      }

      const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
        refreshToken,
      });

      const { accessToken, refreshToken: newRefreshToken, expiresIn } = response.data;

      if (!accessToken || !newRefreshToken) {
        throw new Error('Invalid response: missing tokens');
      }

      // Update stored tokens
      this.setAccessToken(accessToken);
      this.setRefreshToken(newRefreshToken);

      return {
        accessToken,
        refreshToken: newRefreshToken,
        expiresIn,
      };
    } catch (error) {
      // Clear tokens on refresh failure
      this.logout();
      const errorMessage = error.response?.data?.error || error.message || 'Token refresh failed';
      throw new Error(errorMessage);
    }
  }

  /**
   * Logout user - clear all tokens
   */
  logout() {
    this.clearAccessToken();
    this.clearRefreshToken();
  }

  /**
   * Get stored access token
   * @returns {string|null}
   */
  getAccessToken() {
    return localStorage.getItem('accessToken');
  }

  /**
   * Get stored refresh token
   * @returns {string|null}
   */
  getRefreshToken() {
    // Try httpOnly cookie first (if set by backend), fallback to localStorage
    // For now, using localStorage. Backend can set httpOnly cookie in future.
    return localStorage.getItem('refreshToken');
  }

  /**
   * Set access token in localStorage
   * @param {string} token
   */
  setAccessToken(token) {
    localStorage.setItem('accessToken', token);
  }

  /**
   * Set refresh token in localStorage
   * @param {string} token
   */
  setRefreshToken(token) {
    localStorage.setItem('refreshToken', token);
  }

  /**
   * Clear access token
   */
  clearAccessToken() {
    localStorage.removeItem('accessToken');
  }

  /**
   * Clear refresh token
   */
  clearRefreshToken() {
    localStorage.removeItem('refreshToken');
  }

  /**
   * Check if user is authenticated (has access token)
   * @returns {boolean}
   */
  isAuthenticated() {
    return !!this.getAccessToken();
  }
}

export default new AuthService();

