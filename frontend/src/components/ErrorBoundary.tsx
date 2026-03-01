import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
    children?: ReactNode;
}

interface State {
    hasError: boolean;
    error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false
    };

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error('Uncaught error:', error, errorInfo);
    }

    public render() {
        if (this.state.hasError) {
            return (
                <div className="min-h-screen bg-slate-50 flex flex-col items-center justify-center p-6 text-center">
                    <div className="bg-white p-8 rounded-xl shadow-sm max-w-lg w-full border border-red-100">
                        <h1 className="text-2xl font-bold text-slate-800 mb-4">Something went wrong</h1>
                        <p className="text-slate-600 mb-6">
                            An unexpected error occurred in the application. Please try refreshing the page.
                        </p>
                        {this.state.error && (
                            <div className="bg-red-50 p-4 rounded text-left text-sm text-red-800 overflow-auto mb-6">
                                <code>{this.state.error.message}</code>
                            </div>
                        )}
                        <button
                            onClick={() => window.location.reload()}
                            className="px-6 py-2 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition-colors"
                        >
                            Reload Page
                        </button>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}
