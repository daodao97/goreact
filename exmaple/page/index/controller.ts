// Index ub�6h
export const controller = {
  // ubpn�}�p
  async load(params: any, context: any) {
    // !�epn�}
    return {
      title: "Welcome to GoReact",
      message: "This is a server-side rendered React page powered by QuickJS",
      timestamp: new Date().toISOString(),
      userAgent: context?.request?.headers?.['user-agent'] || 'Unknown',
      params: params || {}
    };
  }
};

export default controller;