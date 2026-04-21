import axios, { AxiosInstance } from 'axios';
import type { SearchQuery, SearchResponse, HealthCheck, ServiceStatus, ErrorResponse } from './types';

class DigitalMemoryAPI {
  private client: AxiosInstance;
  private baseURL: string;

  constructor(baseURL: string = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8000') {
    this.baseURL = baseURL;
    this.client = axios.create({
      baseURL: this.baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 30000,
    });
  }

  /**
   * Perform semantic search on stored knowledge
   */
  async search(query: SearchQuery): Promise<SearchResponse> {
    try {
      const response = await this.client.post<SearchResponse>(
        '/api/v1/query',
        {
          query: query.query,
          top_k: query.top_k || 10,
          filters: query.filters || {},
        }
      );
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Get health status of the API
   */
  async health(): Promise<HealthCheck> {
    try {
      const response = await this.client.get<HealthCheck>('/health');
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Get service status and metrics
   */
  async status(): Promise<ServiceStatus> {
    try {
      const response = await this.client.get<ServiceStatus>('/status');
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Get a single knowledge item by ID
   */
  async getKnowledge(id: string) {
    try {
      const response = await this.client.get(`/api/v1/knowledge/${id}`);
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Get all knowledge items with pagination
   */
  async listKnowledge(page: number = 1, limit: number = 20) {
    try {
      const response = await this.client.get('/api/v1/knowledge', {
        params: { page, limit },
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Handle and format API errors
   */
  private handleError(error: any): ErrorResponse {
    if (axios.isAxiosError(error)) {
      const data = error.response?.data as ErrorResponse;
      return {
        error: data?.error || error.message,
        code: error.code || 'UNKNOWN_ERROR',
        details: data?.details,
      };
    }
    return {
      error: 'An unexpected error occurred',
      code: 'UNEXPECTED_ERROR',
    };
  }
}

// Export singleton instance
export const api = new DigitalMemoryAPI();
export default DigitalMemoryAPI;
