package task

import (
	"fmt"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/constant"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/dto"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/entity"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/vo"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/pkg/ibctool"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/repository/cache"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

type ChainOutflowStatisticsTask struct {
	segmentMinTime       int64
	segmentStatisticsMap map[string][]*dto.AggrIBCChainOutflowDTO
}

func (t *ChainOutflowStatisticsTask) Name() string {
	return "ibc_chain_outflow_statistics_task"
}

func (t *ChainOutflowStatisticsTask) Cron() int {
	if taskConf.CronTimeIBCChainOutflowStatisticsTask > 0 {
		return taskConf.CronTimeIBCChainOutflowStatisticsTask
	}
	return EveryHour
}

// Run 增量更新
func (t *ChainOutflowStatisticsTask) Run() int {
	t.todayStatistics()
	t.yesterdayStatistics()
	t.setStatisticsDataCache()
	return 1
}

// RunFullStatistics 全量更新
func (t *ChainOutflowStatisticsTask) RunFullStatistics() int {
	t.segmentMinTime = math.MaxInt64
	t.segmentStatisticsMap = make(map[string][]*dto.AggrIBCChainOutflowDTO)
	segments, err := getTxTimeSegment(false, segmentStepLatest)
	if err != nil {
		logrus.Errorf("task %s getSegment err, %v", t.Name(), err)
		return -1
	}

	historySegments, err := getTxTimeSegment(true, segmentStepHistory)
	if err != nil {
		logrus.Errorf("task %s getHistorySegment err, %v", t.Name(), err)
		return -1
	}

	t.segmentMinTime = segments[0].StartTime

	if err := chainOutflowStatisticsRepo.CreateNew(); err != nil {
		logrus.Errorf("task %s CreateNew err, %v", t.Name(), err)
		return -1
	}

	// 先处理历史表
	logrus.Infof("task %s deal history segment total: %d", t.Name(), len(historySegments))
	t.deal(historySegments, true, true)

	logrus.Infof("task %s deal segment total: %d", t.Name(), len(segments))
	t.deal(segments, false, true)

	if err = chainOutflowStatisticsRepo.SwitchColl(); err != nil {
		logrus.Errorf("task %s SwitchColl err, %v", t.Name(), err)
		return -1
	}

	t.setStatisticsDataCache()
	return 1
}

// deal 对ibc tx表的数据进行统计
//	- targetHistory true: 统计ex_ibc_tx表; false: 统计ex_ibc_tx_latest表
//  - fullStatistics true: 统计数据写入新表(xxx_new); 当全量统计时，此值为true
func (t *ChainOutflowStatisticsTask) deal(segments []*segment, targetHistory bool, fullStatistics bool) {
	for _, v := range segments {
		logrus.Infof("task %s deal segment [%d, %d], targetHistory: %t", t.Name(), v.StartTime, v.EndTime, targetHistory)

		aggrRes, err := ibcTxRepo.AggrIBCChainOutflow(v.StartTime, v.EndTime, targetHistory)
		if err != nil {
			logrus.Errorf("task %s AggrIBCChainOutflow segment [%d, %d], targetHistory: %t err, %v", t.Name(), v.StartTime, v.EndTime, targetHistory, err)
			continue
		}

		if len(aggrRes) == 0 {
			continue
		}

		if fullStatistics {
			aggrRes = t.integrationStatisticsData(aggrRes, v, targetHistory)
		}

		if err = t.saveData(aggrRes, v, targetHistory, fullStatistics); err != nil {
			logrus.Errorf("task %s dealHistory saveData err, %v", t.Name(), err)
		}
	}
}

func (t *ChainOutflowStatisticsTask) integrationStatisticsData(aggrRes []*dto.AggrIBCChainOutflowDTO, seg *segment, targetHistory bool) []*dto.AggrIBCChainOutflowDTO {
	// 将历史表与新表的重叠的分段记录下来
	if targetHistory {
		if seg.StartTime >= t.segmentMinTime {
			t.segmentStatisticsMap[fmt.Sprintf("%d-%d", seg.StartTime, seg.EndTime)] = aggrRes
			return aggrRes
		}
	}

	// 新表中的段与历史表重和，需要整合数据
	hirtoryAggrRes, ok := t.segmentStatisticsMap[fmt.Sprintf("%d-%d", seg.StartTime, seg.EndTime)]
	if !ok {
		return aggrRes
	}

	integrationDataMap := make(map[string]*dto.AggrIBCChainOutflowDTO, len(aggrRes))
	for _, v := range aggrRes {
		key := fmt.Sprintf("%s%s%s%d", v.Chain, v.BaseDenomChain, v.BaseDenom, v.Status)
		if data, ok := integrationDataMap[key]; !ok {
			integrationDataMap[key] = v
		} else {
			data.TxsNum += v.TxsNum
			data.DenomAmount += v.DenomAmount
		}
	}

	for _, v := range hirtoryAggrRes {
		key := fmt.Sprintf("%s%s%s%d", v.Chain, v.BaseDenomChain, v.BaseDenom, v.Status)
		if data, ok := integrationDataMap[key]; !ok {
			integrationDataMap[key] = v
		} else {
			data.TxsNum += v.TxsNum
			data.DenomAmount += v.DenomAmount
		}
	}

	integrationDataList := make([]*dto.AggrIBCChainOutflowDTO, 0, len(integrationDataMap))
	for _, v := range integrationDataMap {
		integrationDataList = append(integrationDataList, v)
	}
	return integrationDataList
}

func (t *ChainOutflowStatisticsTask) saveData(aggrRes []*dto.AggrIBCChainOutflowDTO, seg *segment, targetHistory bool, fullStatistics bool) error {
	nowTime := time.Now().Unix()
	entityList := make([]*entity.IBCChainOutflowStatistics, 0, len(aggrRes))

	for _, v := range aggrRes {
		entityList = append(entityList, &entity.IBCChainOutflowStatistics{
			Chain:            v.Chain,
			BaseDenom:        v.BaseDenom,
			BaseDenomChain:   v.BaseDenomChain,
			Status:           entity.IbcTxStatus(v.Status),
			DenomAmount:      v.DenomAmount,
			TxsNumber:        v.TxsNum,
			SegmentStartTime: seg.StartTime,
			SegmentEndTime:   seg.EndTime,
			CreateAt:         nowTime,
			UpdateAt:         nowTime,
		})
	}

	var err error
	if fullStatistics {
		if targetHistory {
			err = chainOutflowStatisticsRepo.InsertManyToNew(entityList)
		} else {
			err = chainOutflowStatisticsRepo.BatchSwapNew(seg.StartTime, seg.EndTime, entityList)
		}
	} else {
		err = chainOutflowStatisticsRepo.BatchSwap(seg.StartTime, seg.EndTime, entityList)
	}

	return err
}

func (t *ChainOutflowStatisticsTask) todayStatistics() {
	logrus.Infof("task %s exec today statistics", t.Name())
	startTime, endTime := todayUnix()
	segments := []*segment{
		{
			StartTime: startTime,
			EndTime:   endTime,
		},
	}

	t.deal(segments, false, false)
}

func (t *ChainOutflowStatisticsTask) yesterdayStatistics() {
	ok, seg := whetherCheckYesterdayStatistics(t.Name(), t.Cron())
	if !ok {
		return
	}

	logrus.Infof("task %s check yeaterday statistics", t.Name())
	t.deal([]*segment{seg}, false, false)
}

func (t *ChainOutflowStatisticsTask) setStatisticsDataCache() {
	days := constant.ChainFlowTrendDays
	startTime, _ := lastNDaysZeroTimeUnix(days)
	_, endTime := todayUnix()

	chainInfosMap, err := getAllChainInfosMap()
	if err != nil {
		logrus.Errorf("task %s getAllChainInfosMap err, %v", t.Name(), err)
		return
	}

	priceMap := cache.TokenPriceMap()
	for chain, _ := range chainInfosMap {
		trendList, err := chainOutflowStatisticsRepo.AggrTrend(chain, startTime, endTime)
		if err != nil {
			logrus.Errorf("task %s AggrTrend %s err, %v", t.Name(), chain, err)
			continue
		}

		volumeMap := make(map[string]decimal.Decimal, len(trendList))
		totalDenomValue := decimal.Zero
		for _, v := range trendList {
			denomAmount := decimal.NewFromFloat(v.DenomAmount)
			denomValue := ibctool.CalculateDenomValue(priceMap, v.BaseDenom, v.BaseDenomChain, denomAmount)
			dt := time.Unix(v.SegmentStartTime, 0).Format(constant.DateFormat)
			if vol, ok := volumeMap[dt]; ok {
				volumeMap[dt] = vol.Add(denomValue)
			} else {
				volumeMap[dt] = denomValue
			}
			totalDenomValue = totalDenomValue.Add(denomValue)
		}

		volumeItemList := make([]vo.VolumeItem, 0, len(volumeMap))
		for dt, vol := range volumeMap {
			volumeItemList = append(volumeItemList, vo.VolumeItem{
				Datetime: dt,
				Value:    vol.String(),
			})
		}

		if err = chainFlowCacheRepo.SetOutflowTrend(days, chain, volumeItemList); err != nil {
			logrus.Errorf("task %s SetOutflowTrend %s err, %v", t.Name(), chain, err)
		}

		if err = chainFlowCacheRepo.SetOutflowVolume(days, chain, totalDenomValue.String()); err != nil {
			logrus.Errorf("task %s SetOutflowVolume %s err, %v", t.Name(), chain, err)
		}
	}

	chainFlowCacheRepo.ExpireOutflowTrend(days, OneWeek*time.Second)
	chainFlowCacheRepo.ExpireOutflowVolume(days, OneWeek*time.Second)
}
