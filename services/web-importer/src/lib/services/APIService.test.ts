/// <reference types="bun-types/test-globals" />
import { test, expect, describe, jest, afterEach, beforeEach } from "bun:test";
import { collect } from "./APIService";

// --- Code to Test ---
// (Assuming this is imported from './collect.ts')
// I'm including it here for a self-contained example.

// Helper types (assuming definitions based on the function)
declare namespace API {
  export interface PaginatedResponse {
    links?: {
      next?: string | null;
    };
  }
}

// This is a guess at the RequestResult type based on the function's usage
type RequestResult<TData, TError, ThrowOnError> =
  | { data: TData; error?: undefined }
  | { data: undefined; error: TError };

// --- Test Setup ---

// Define a concrete type for our tests
interface Device {
  id: number;
  name: string;
}

// Define concrete response types for our mock
type DeviceResponse = API.PaginatedResponse & { data: Device[] };
type PaginatedData = Record<number, DeviceResponse>;
type PaginatedError = Record<number, { message: string }>;

// Mock the non-standard URL.parse function
// We'll make it behave like `new URL()` for the test
const mockUrlParse = jest.fn((url: string) => {
  // Handle the '?? ""' case, which would throw in `new URL()`
  if (!url) {
    return { searchParams: new URLSearchParams() };
  }
  // Use the standard URL object to parse the URL string
  return new URL(url);
});

describe("collect", () => {
  // Spy on the global URL object to mock 'parse'
  let urlParseSpy: any;

  beforeEach(() => {
    // Mock URL.parse before each test
    urlParseSpy = jest
      .spyOn(globalThis.URL, "parse")
      .mockImplementation(mockUrlParse as any);
  });

  afterEach(() => {
    // Restore the original implementation and clear mocks
    urlParseSpy.mockRestore();
    jest.clearAllMocks();
  });

  // Test 1: Single page of data
  test("should collect data from a single page", async () => {
    const page1Data: Device[] = [{ id: 1, name: "Device A" }];
    const mockCall = jest.fn().mockResolvedValue({
      data: {
        data: page1Data,
        links: { next: null }, // No next page
      },
    });

    const result = await collect(
      mockCall as any,
    );

    expect(result).toEqual(page1Data);
    expect(mockCall).toHaveBeenCalledTimes(1);
    expect(mockCall).toHaveBeenCalledWith(undefined); // First call has no cursor
  });

  // Test 2: Multiple pages of data
  test("should collect and concatenate data from multiple pages", async () => {
    const page1Data: Device[] = [{ id: 1, name: "Device A" }];
    const page2Data: Device[] = [{ id: 2, name: "Device B" }];

    const mockCall = jest
      .fn()
      // First call response
      .mockResolvedValueOnce({
        data: {
          data: page1Data,
          links: { next: "https://api.example.com/devices?cursor=page2" },
        },
      })
      // Second call response
      .mockResolvedValueOnce({
        data: {
          data: page2Data,
          links: { next: null }, // No more pages
        },
      });

    const result = await collect(
      mockCall as any,
    );

    // Check final concatenated data
    expect(result).toEqual([
      { id: 1, name: "Device A" },
      { id: 2, name: "Device B" },
    ]);

    // Check that the mock was called correctly
    expect(mockCall).toHaveBeenCalledTimes(2);
    expect(mockCall).toHaveBeenNthCalledWith(1, undefined);
    expect(mockCall).toHaveBeenNthCalledWith(2, "page2");

    // Check that our URL.parse mock was used
    expect(mockUrlParse).toHaveBeenCalledWith(
      "https://api.example.com/devices?cursor=page2",
    );
  });

  // Test 3: Empty response
  test("should return an empty array if no data is found", async () => {
    const mockCall = jest.fn().mockResolvedValue({
      data: {
        data: [], // Empty data array
        links: { next: null },
      },
    });

    const result = await collect(
      mockCall as any,
    );

    expect(result).toEqual([]);
    expect(mockCall).toHaveBeenCalledTimes(1);
  });

  // Test 4: API Error
  test("should return an Error if the API call fails", async () => {
    const apiError = { message: "Internal Server Error" };
    const mockCall = jest.fn().mockResolvedValue({
      data: undefined,
      error: apiError,
    });

    const result = await collect(
      mockCall as any,
    );

    expect(result).toBeInstanceOf(Error);
    expect(mockCall).toHaveBeenCalledTimes(1);
  });

  // Test 5: Gracefully handles empty 'next' string
  test("should stop paginating if 'next' link is an empty string", async () => {
    const page1Data: Device[] = [{ id: 1, name: "Device A" }];
    const mockCall = jest.fn().mockResolvedValue({
      data: {
        data: page1Data,
        links: { next: "" }, // Empty string
      },
    });

    const result = await collect(
      mockCall as any,
    );

    expect(result).toEqual(page1Data);
    expect(mockCall).toHaveBeenCalledTimes(1);
    expect(mockUrlParse).toHaveBeenCalledWith(""); // Our mock handles this
  });
});
