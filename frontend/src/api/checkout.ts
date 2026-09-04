import { API_BASE_URL, authFetch } from "./config";

// Types
export interface CheckoutItem {
  product_id: string;
  product_name: string;
  quantity: number;
  unit_price: number;
  total_price: number;
}

export interface DeliveryInfo {
  customer_name: string;
  customer_phone: string;
  street: string;
  house_number: string;
  postal_code: string;
  city: string;
  floor?: string;
  instructions?: string;
}

export interface CheckoutRequest {
  user_id: string;
  items: CheckoutItem[];
  total_amount: number;
  currency: string;
  order_type: "delivery" | "pickup";
  delivery_info?: DeliveryInfo;
}

export interface CheckoutResponse {
  status: string;
  order_id: string;
  order_type: "delivery" | "pickup";
  total_amount: number;
  currency: string;
  created_at: string;
}

export interface PaymentRequest {
  order_id: string;
  user_id: string;
  provider: "stripe" | "paypal" | "bank_transfer";
  amount: number;
  currency: string;
  payment_details:
    | StripePaymentDetails
    | PayPalPaymentDetails
    | BankTransferDetails;
}

export interface StripePaymentDetails {
  card_number: string;
  expiry_month: string;
  expiry_year: string;
  cvv: string;
  cardholder_name: string;
}

export interface PayPalPaymentDetails {
  email: string;
  password: string;
}

export interface BankTransferDetails {
  iban: string;
  account_holder: string;
  bank_name?: string;
}

export interface PaymentResponse {
  id: string;
  order_id: string;
  status: "pending" | "success" | "failed";
}

export interface OrderStatus {
  order_id: string;
  status: string;
  items?: string[];
  created_at?: string;
}

export interface DeliveryStatus {
  order_id: string;
  order_received: boolean;
  kitchen_ready: boolean;
  delivery_started: boolean;
  order_type: "delivery" | "pickup";
  status: string;
  delivery_info?: DeliveryInfo;
  driver_id?: string;
  created_at: string;
  updated_at: string;
}

export const checkoutApi = {
  async checkout(request: CheckoutRequest): Promise<CheckoutResponse> {
    const response = await authFetch(`${API_BASE_URL}/shopping/checkout`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Checkout failed");
    }
    return response.json();
  },

  async processPayment(request: PaymentRequest): Promise<PaymentResponse> {
    const response = await authFetch(`${API_BASE_URL}/payment/payments`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Payment failed");
    }
    return response.json();
  },

  async getOrderStatus(orderId: string): Promise<OrderStatus> {
    const response = await authFetch(`${API_BASE_URL}/kitchen/orders/${orderId}`);
    if (!response.ok) throw new Error("Failed to fetch order status");
    return response.json();
  },

  async getDeliveryStatus(orderId: string): Promise<DeliveryStatus | null> {
    try {
      const response = await fetch(`${API_BASE_URL}/delivery/orders/${orderId}`);
      if (!response.ok) return null;
      return response.json();
    } catch {
      return null;
    }
  },
};
