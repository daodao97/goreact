// Index ubpn!‹
export interface IndexPageData {
  title: string;
  message: string;
  timestamp: string;
  userAgent: string;
  params: Record<string, any>;
}

export interface IndexPageProps {
  data: IndexPageData;
}

export default IndexPageData;