import { FaqSection } from "@/components/sections/FaqSection";
import { GetStartedSection } from "@/components/sections/GetStartedSection";
import { HeroSection } from "@/components/sections/HeroSection";
import { HowItWorksSection } from "@/components/sections/HowItWorksSection";
import { SiteFooter } from "@/components/sections/SiteFooter";
import { SiteHeader } from "@/components/sections/SiteHeader";
import { WhoItIsForSection } from "@/components/sections/WhoItIsForSection";
import { WhyFgmSection } from "@/components/sections/WhyFgmSection";
import { GridBackground } from "@/components/ui/GridBackground";

export default function HomePage() {
  return (
    <>
      <GridBackground />
      <a className="skip-link" href="#main-content">
        Skip to main content
      </a>
      <div
        id="copy-live-status"
        className="sr-only"
        role="status"
        aria-live="polite"
      />

      <SiteHeader />

      <main id="main-content" className="page-main">
        <HeroSection />
        <WhyFgmSection />
        <HowItWorksSection />
        <GetStartedSection />
        <WhoItIsForSection />
        <FaqSection />
      </main>

      <SiteFooter />
    </>
  );
}
