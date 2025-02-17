import React from "react";

import LinkArrow from "../../../../../assets/images/icon-arrow-right-vibrant-blue-10x18@2x.png";

const baseClass = "learn-fleet";

const LearnFleet = (): JSX.Element => {
  return (
    <div className={baseClass}>
      <p>
        Want to explore Fleet&apos;s features? Learn how to ask questions about
        your device using queries.
      </p>
      <a
        className="dashboard-info-card__action-button"
        href="https://fleetdm.com/docs/using-fleet/learn-how-to-use-fleet"
        target="_blank"
        rel="noopener noreferrer"
      >
        Learn how to use Fleet&nbsp;
        <img src={LinkArrow} alt="link arrow" id="link-arrow" />
      </a>
    </div>
  );
};

export default LearnFleet;
