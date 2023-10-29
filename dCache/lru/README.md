# LRU

### FIFO / LFU / LRU
- FIFO (First In First Out)
  - 近考虑时间因素
  - 会导致缓存命中率太低
- LFU (Least Frequently Used)
  - 仅考虑访问频率
  - 淘汰访频率最低的数据 
  - 如果历史上某数据访问频率很高，突然不再访问了，但是由于之前访问频率很高，LFU 算法可能会一直保存这个数据
- LRU (Least Recently Used)
  - 考虑时间和访问频率 
  - 淘汰最近最少使用的数据
  - 优点：简单、高效、容易实现
  - 缺点：需要维护访问历史，访问历史的维护成本高

### LRU core
- 字典 + 双向链表
![LRU 核心数据结构](../../assert/lru.jpg)