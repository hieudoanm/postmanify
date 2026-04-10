'use client';

import Link from 'next/link';
import { FC, useEffect, useState } from 'react';

export const ErrorTemplate: FC<{
  error: { code: number; message: string };
  messages: string[];
}> = ({
  error = { code: 500, message: 'Internal Server Error' },
  messages = [],
}) => {
  const [{ message = '', path = '' }, setState] = useState<{
    message: string;
    path: string;
  }>({
    message: '',
    path: '',
  });

  useEffect(() => {
    const randomizeMessage = () => {
      const randomMessage: string =
        messages[Math.floor(Math.random() * messages.length)];
      const notFoundPath: string =
        typeof window === 'undefined' ? '' : window.location.pathname;
      setState({ message: randomMessage, path: notFoundPath });
    };
    randomizeMessage();
  }, [messages]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Backspace' || e.key === 'ArrowLeft') {
        window.history.back();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  const isNotFound = error.code === 404;

  return (
    <div
      className="bg-base-100 text-base-content relative flex min-h-screen items-center justify-center px-12 py-24"
      data-theme="luxury">
      {/* Ambient glow — mirrors Hero's radial blob */}
      <div className="bg-primary/5 pointer-events-none absolute top-1/2 left-1/2 h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 rounded-full blur-3xl" />

      <div className="relative mx-auto w-full max-w-xl text-center">
        {/* Error code badge */}
        <p className="text-primary mb-7 text-xs tracking-[0.2em] uppercase">
          {isNotFound ? 'Page not found' : 'Something went wrong'}
        </p>

        {/* Giant error code — mirrors Hero h1 treatment */}
        <h1 className="mb-6 font-serif text-[9rem] leading-none font-black tracking-tight">
          <span className="text-primary">{error.code}</span>
        </h1>

        {/* Message */}
        <h2 className="mb-4 font-serif text-3xl leading-snug font-bold">
          {error.message}
        </h2>

        {/* Random flavour copy */}
        {message && (
          <p className="text-base-content/60 mx-auto mb-3 max-w-sm text-base leading-relaxed">
            {message}
          </p>
        )}

        {/* Requested path */}
        {path && (
          <p className="text-base-content/30 mb-10 font-mono text-xs break-all">
            {path}
          </p>
        )}

        {/* CTA buttons — mirrors Hero button row */}
        <div className="mb-10 flex flex-wrap justify-center gap-3">
          <Link href="/" className="btn btn-primary">
            Go home
          </Link>
          <button
            className="btn btn-ghost"
            onClick={() => window.history.back()}>
            Go back
          </button>
        </div>

        {/* Keyboard hint — mirrors Hero badges row */}
        <div className="flex flex-wrap justify-center gap-3">
          <span className="badge badge-ghost border-base-300 text-base-content/40 gap-1.5 border">
            <kbd className="kbd kbd-xs">←</kbd>
            <span>or</span>
            <kbd className="kbd kbd-xs">Backspace</kbd>
            <span>to go back</span>
          </span>
          {isNotFound && (
            <span className="badge badge-warning text-warning-content">
              404
            </span>
          )}
          {!isNotFound && (
            <span className="badge badge-error">{error.code}</span>
          )}
        </div>
      </div>
    </div>
  );
};

ErrorTemplate.displayName = 'ErrorTemplate';
