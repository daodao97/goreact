import React from 'react';
import { IndexPageProps } from './model';

// Index ubÆþÄö
export const IndexPage: React.FC<IndexPageProps> = ({ data }) => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-xl shadow-lg p-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-6">{data.title}</h1>
          
          <div className="space-y-4">
            <div className="p-4 bg-blue-50 rounded-lg">
              <h2 className="text-lg font-semibold text-blue-900 mb-2">Server Message</h2>
              <p className="text-blue-700">{data.message}</p>
            </div>
            
            <div className="grid md:grid-cols-2 gap-4">
              <div className="p-4 bg-green-50 rounded-lg">
                <h3 className="font-semibold text-green-900 mb-2">Timestamp</h3>
                <p className="text-green-700 text-sm font-mono">{data.timestamp}</p>
              </div>
              
              <div className="p-4 bg-purple-50 rounded-lg">
                <h3 className="font-semibold text-purple-900 mb-2">User Agent</h3>
                <p className="text-purple-700 text-xs font-mono break-all">{data.userAgent}</p>
              </div>
            </div>
            
            {Object.keys(data.params).length > 0 && (
              <div className="p-4 bg-yellow-50 rounded-lg">
                <h3 className="font-semibold text-yellow-900 mb-2">Route Parameters</h3>
                <pre className="text-yellow-700 text-sm">
                  {JSON.stringify(data.params, null, 2)}
                </pre>
              </div>
            )}
          </div>
          
          <div className="mt-8 text-center">
            <p className="text-gray-600">
              Powered by <strong>Go + React + QuickJS</strong>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default IndexPage;