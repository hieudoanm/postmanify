'use client';

import { FC, useState } from 'react';

export const VersionTemplate: FC<{
  version: string;
}> = ({ version }) => {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    await navigator.clipboard.writeText(version);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  // Parse version segments for display (YYYY.MM.DD.hh.mm.ss)
  const [year, month, day, hh, mm, ss] = version.split('.');
  const hasSegments = year && month && day;

  return (
    <div
      className="bg-base-100 text-base-content relative flex min-h-screen items-center justify-center px-12 py-24"
      data-theme="luxury">
      {/* Ambient glow — mirrors Hero radial blob */}
      <div className="bg-primary/5 pointer-events-none absolute top-1/2 left-1/2 h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 rounded-full blur-3xl" />

      <div className="relative mx-auto w-full max-w-lg text-center">
        {/* Section label — mirrors LandingTemplate Section label pattern */}
        <p className="text-primary mb-7 text-xs tracking-[0.2em] uppercase">
          Current deployment
        </p>

        {/* Heading */}
        <h1 className="mb-4 font-serif text-5xl leading-[1.05] font-black tracking-tight">
          App Version
        </h1>

        <p className="text-base-content/60 mx-auto mb-11 max-w-sm text-base leading-relaxed">
          Build version of the current deployment
        </p>

        {/* Version display card — mirrors DataSection card style */}
        <div className="card bg-base-200 border-base-300 mb-8 border">
          <div className="card-body p-7">
            {hasSegments ? (
              <div className="flex items-center justify-center gap-0">
                {/* Date group */}
                <div className="flex flex-col items-center px-5">
                  <span className="text-primary font-mono text-3xl font-bold">
                    {year}
                  </span>
                  <span className="text-base-content/30 mt-1 text-[11px] tracking-widest uppercase">
                    Year
                  </span>
                </div>
                <span className="text-base-content/20 font-mono text-2xl">
                  .
                </span>
                <div className="flex flex-col items-center px-5">
                  <span className="font-mono text-3xl font-bold">{month}</span>
                  <span className="text-base-content/30 mt-1 text-[11px] tracking-widest uppercase">
                    Month
                  </span>
                </div>
                <span className="text-base-content/20 font-mono text-2xl">
                  .
                </span>
                <div className="flex flex-col items-center px-5">
                  <span className="font-mono text-3xl font-bold">{day}</span>
                  <span className="text-base-content/30 mt-1 text-[11px] tracking-widest uppercase">
                    Day
                  </span>
                </div>
                {hh && (
                  <>
                    <span className="text-base-content/20 font-mono text-2xl">
                      .
                    </span>
                    <div className="flex flex-col items-center px-5">
                      <span className="font-mono text-3xl font-bold">{hh}</span>
                      <span className="text-base-content/30 mt-1 text-[11px] tracking-widest uppercase">
                        Hour
                      </span>
                    </div>
                  </>
                )}
                {mm && (
                  <>
                    <span className="text-base-content/20 font-mono text-2xl">
                      .
                    </span>
                    <div className="flex flex-col items-center px-5">
                      <span className="font-mono text-3xl font-bold">{mm}</span>
                      <span className="text-base-content/30 mt-1 text-[11px] tracking-widest uppercase">
                        Min
                      </span>
                    </div>
                  </>
                )}
                {ss && (
                  <>
                    <span className="text-base-content/20 font-mono text-2xl">
                      .
                    </span>
                    <div className="flex flex-col items-center px-5">
                      <span className="font-mono text-3xl font-bold">{ss}</span>
                      <span className="text-base-content/30 mt-1 text-[11px] tracking-widest uppercase">
                        Sec
                      </span>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <p className="text-primary font-mono text-2xl font-bold break-all">
                {version}
              </p>
            )}
          </div>
        </div>

        {/* Actions — mirrors Hero CTA row */}
        <div className="mb-10 flex flex-wrap justify-center gap-3">
          <button
            onClick={copy}
            className={`btn ${copied ? 'btn-success' : 'btn-primary'}`}>
            {copied ? '✓ Copied' : 'Copy version'}
          </button>
          <button className="btn btn-ghost font-mono text-sm" onClick={copy}>
            {version}
          </button>
        </div>

        {/* Format hint — mirrors Hero badges row */}
        <div className="flex flex-wrap justify-center gap-3">
          <span className="badge badge-ghost border-base-300 text-base-content/40 gap-1.5 border">
            Format:
            <kbd className="kbd kbd-xs">YYYY.MM.DD.hh.mm.ss</kbd>
          </span>
          <span className="badge badge-neutral">Stable</span>
        </div>
      </div>
    </div>
  );
};

VersionTemplate.displayName = 'VersionTemplate';
