/**
 * step: the interval between each point unit:s  for example: half an hour 1800s, interval 30s, return sixty points
 * @param pointNumber points，default 15
 * @param startTimeStamp unit:s
 * @param endTimeStamp
 * @returns
 */
export const caculateStep = (
  startTimeStamp: number,
  endTimeStamp: number,
  pointNumber: number
): number => {
  return Math.ceil((endTimeStamp - startTimeStamp) / pointNumber);
};
